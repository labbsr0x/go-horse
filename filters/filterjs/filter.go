package filterjs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/model"

	"github.com/kataras/iris/core/errors"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/plugins"
	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"

	"github.com/robertkrimen/otto"
)

var client = &http.Client{}

// FilterJS JS proxy filter
type FilterJS struct {
	model.FilterConfig
}

// NewFilterJS JS filter factory
func NewFilterJS(innerType model.FilterConfig) FilterJS {
	filterJs := FilterJS{}
	filterJs.FilterConfig = innerType
	return filterJs
}

// MatchURL js
func (filterJs FilterJS) MatchURL(ctx iris.Context) bool {
	return filterJs.Regex.MatchString(ctx.RequestPath(false))
}

// Config js
func (filterJs FilterJS) Config() model.FilterConfig {
	return filterJs.FilterConfig
}

// Exec run the filter
func (filterJs FilterJS) Exec(ctx iris.Context, body string) (model.FilterReturn, error) {
	js := otto.New()

	emptyBody, _ := js.Object("({})")
	bodyParsed, _ := otto.ToValue(emptyBody)
	var error error
	var contentType string
	var headers http.Header
	if filterJs.Invoke == model.Request {
		contentType = ctx.Request().Header.Get("Content-Type")
		headers = ctx.Request().Header
	} else {
		contentType = ctx.ResponseWriter().Header().Get("Content-Type")
		headers = ctx.ResponseWriter().Header()
	}
	if contentType == "application/json" {
		if body == "" {
			body = "{}"
		}
		bodyParsed, error = js.Call("JSON.parse", nil, body)
		if error != nil {
			log.Error().Str("plugin_name", filterJs.Name).Err(error).Msg("Error parsing body string to JS object - js filter exec")
			bodyParsed, _ = otto.ToValue(emptyBody)
		}
	}

	operation, error := js.Object("({READ : 0, WRITE : 1})")
	if error != nil {
		log.Error().Str("plugin_name", filterJs.Name).Err(error).Msg("Error creating operation object - js filter exec")
	}

	urlParamsJsObj, _ := js.Object("({})")
	urlParamsJsObj.Set("add", func(call otto.FunctionCall) otto.Value { return addURLParamsFromJSContext(ctx, call) })
	urlParamsJsObj.Set("set", func(call otto.FunctionCall) otto.Value { return setURLParamsFromJSContext(ctx, call) })
	urlParamsJsObj.Set("get", func(call otto.FunctionCall) otto.Value { return getURLParamsFromJSContext(ctx, call) })
	urlParamsJsObj.Set("del", func(call otto.FunctionCall) otto.Value { return delURLParamsFromJSContext(ctx, call) })
	urlParamsJsObj.Set("list", func(call otto.FunctionCall) otto.Value { return listURLParamsFromJSContext(ctx, call) })

	valuesJsObj, _ := js.Object("({})")
	valuesJsObj.Set("get", func(call otto.FunctionCall) otto.Value { return requestScopeGetToJSContext(ctx, call) })
	valuesJsObj.Set("set", func(call otto.FunctionCall) otto.Value { return requestScopeSetToJSContext(ctx, call) })
	valuesJsObj.Set("list", func(call otto.FunctionCall) otto.Value { return requestScopeListToJSContext(ctx, call) })

	ctxJsObj, _ := js.Object("({})")
	ctxJsObj.Set("url", ctx.Request().URL.Path)
	ctxJsObj.Set("body", bodyParsed.Object())
	ctxJsObj.Set("operation", operation)
	ctxJsObj.Set("method", strings.ToUpper(ctx.Method()))
	ctxJsObj.Set("headers", headers)
	ctxJsObj.Set("request", httpRequestTOJSContext)
	ctxJsObj.Set("values", valuesJsObj)
	ctxJsObj.Set("urlParams", urlParamsJsObj)
	ctxJsObj.Set("responseStatusCode", ctx.Values().GetString("responseStatusCode"))

	js.Set("ctx", ctxJsObj)

	pluginsJsObj, _ := js.Object("({})")

	for _, jsPlugin := range plugins.JSPluginList {
		error := pluginsJsObj.Set(jsPlugin.Name(), func(call otto.FunctionCall) otto.Value { return jsPlugin.Set(ctx, call) })
		if error != nil {
			log.Error().Str("plugin_name", jsPlugin.Name()).Err(error).Msg("Error on applying GO->JS plugin - js filter exec")
		}
	}

	js.Set("plugins", pluginsJsObj)

	returnValue, error := js.Run("(" + filterJs.Function + ")(ctx, plugins)")

	if error != nil {
		log.Error().Str("plugin_name", filterJs.Name).Err(error).Msg("Error executing filter - js filter exec")
		return model.FilterReturn{Next: false, Body: "{\"message\" : \"Error from docker daemon proxy go-horse : \"" + error.Error() + "}"}, error
	}

	result := returnValue.Object()

	jsFunctionReturn := model.FilterReturn{}

	if value, err := result.Get("next"); err == nil {
		if value, err := value.ToBoolean(); err == nil {
			jsFunctionReturn.Next = value
		} else {
			return errorReturnFilter(error)
		}
	} else {
		return errorReturnFilter(error)
	}

	if value, err := result.Get("body"); err == nil {
		if value, err := js.Call("JSON.stringify", nil, value); err == nil {
			jsFunctionReturn.Body = value.String()
		} else {
			return errorReturnFilter(error)
		}
	} else {
		return errorReturnFilter(error)
	}

	if value, err := result.Get("operation"); err == nil {
		if value, err := value.ToInteger(); err == nil {
			jsFunctionReturn.Operation = model.BodyOperation(value)
		} else {
			return errorReturnFilter(error)
		}
	} else {
		return errorReturnFilter(error)
	}

	if value, err := result.Get("status"); err == nil {
		if value, err := value.ToInteger(); err == nil {
			jsFunctionReturn.Status = int(value)
		} else {
			return errorReturnFilter(error)
		}
	} else {
		return errorReturnFilter(error)
	}

	if value, err := result.Get("error"); err == nil {
		if value.IsDefined() {
			if value, err := value.ToString(); err == nil {
				jsFunctionReturn.Err = errors.New(value)
			} else {
				return errorReturnFilter(error)
			}
		}
	} else {
		return errorReturnFilter(error)
	}
	// weird
	return jsFunctionReturn, jsFunctionReturn.Err
}

func errorReturnFilter(error error) (model.FilterReturn, error) {
	log.Error().Err(error).Msg("Error parsing filter return value - js filter exec")
	return model.FilterReturn{Body: "{\"message\" : \"Proxy error : \"" + error.Error() + "}"}, error
}

func httpRequestTOJSContext(call otto.FunctionCall) otto.Value {
	method, error := call.Argument(0).ToString()
	if error != nil {
		log.Error().Err(error).Msg("Error parsing httpRequestTOJSContext method - js filter exec")
	}
	url, error := call.Argument(1).ToString()
	if error != nil {
		log.Error().Err(error).Msg("Error parsing httpRequestTOJSContext url - js filter exec")
	}
	body, error := call.Argument(2).ToString()
	if error != nil {
		log.Error().Err(error).Msg("Error parsing httpRequestTOJSContext body - js filter exec")
	}
	var req *http.Request
	var err interface{}

	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			log.Error().Err(error).Msg("Error parsing httpRequestTOJSContext GET header - js filter exec")
		}
	} else {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
		if err != nil {
			log.Error().Err(error).Msg("Error parsing httpRequestTOJSContext OTHER THAN GET header - js filter exec")
		}
	}

	headers := call.Argument(3).Object()
	if headers != nil {
		for _, key := range headers.Keys() {
			header, error := headers.Get(key)
			if error != nil {
				log.Error().Err(error).Msg("Error parsing httpRequestTOJSContext GET header - js filter exec")
			}
			headerValue, error := header.ToString()
			if error != nil {
				log.Error().Err(error).Msg("Error parsing httpRequestTOJSContext GET header - js filter exec")
			}
			req.Header.Add(key, headerValue)
		}
	}
	log.Debug().Str("method", method).Str("url", url).Str("body", body).Str("headers", fmt.Sprintf("%#v", headers)).Msg("Request parameters")
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Msg("Error executing the request - httpRequestTOJSContext " + fmt.Sprintf("%#v", err))
		response, _ := call.Otto.Object("({})")
		buf, marshalError := json.Marshal(err)
		if marshalError == nil {
			response.Set("body", fmt.Sprintf("%v", string(buf)))
		} else {
			response.Set("body", fmt.Sprintf("%#v", err))
		}
		response.Set("status", 0)
		value, _ := otto.ToValue(response)
		return value
	}
	defer resp.Body.Close()
	if req.Body != nil {
		defer req.Body.Close()
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Msg("Error parsing request body - httpRequestTOJSContext " + fmt.Sprintf("%#v", err))
	}

	response, _ := call.Otto.Object("({})")

	bodyObjectJs, error := call.Otto.ToValue(string(bodyBytes))
	if error != nil {
		log.Error().Err(error).Msg("Error parsing response body to JS object - httpRequestTOJSContext")
	}

	headersObjectJs, error := call.Otto.ToValue(resp.Header)
	if error != nil {
		log.Error().Err(error).Msg("Error parsing response headers to JS object - httpRequestTOJSContext")
	}

	response.Set("body", bodyObjectJs)
	response.Set("status", resp.StatusCode)
	response.Set("headers", headersObjectJs)

	value, _ := otto.ToValue(response)

	return value
}
