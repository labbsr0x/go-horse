package filterjs

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labbsr0x/go-horse/filters/model"

	"github.com/kataras/iris/core/errors"

	"github.com/labbsr0x/go-horse/plugins"
	"github.com/kataras/iris"
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
	var err error
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
		bodyParsed, err = js.Call("JSON.parse", nil, body)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"plugin_name": filterJs.Name,
			}).Errorf("Error parsing body string to JS object - js filter exec")
			bodyParsed, _ = otto.ToValue(emptyBody)
		}
	}

	operation, err := js.Object("({READ : 0, WRITE : 1})")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"plugin_name": filterJs.Name,
		}).Errorf("Error creating operation object - js filter exec")
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
		err := pluginsJsObj.Set(jsPlugin.Name(), func(call otto.FunctionCall) otto.Value { return jsPlugin.Set(ctx, call) })
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"plugin_name": jsPlugin.Name(),
			}).Errorf("Error on applying GO->JS plugin - js filter exec")
		}
	}

	js.Set("plugins", pluginsJsObj)

	returnValue, err := js.Run("(" + filterJs.Function + ")(ctx, plugins)")

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"plugin_name": filterJs.Name,
		}).Errorf("plugin_name: %v | Error executing filter - js filter exec")
		return model.FilterReturn{Next: false, Body: "{\"message\" : \"Error from docker daemon proxy go-horse : \"" + err.Error() + "}"}, err
	}

	result := returnValue.Object()

	jsFunctionReturn := model.FilterReturn{}

	if value, err := result.Get("next"); err == nil {
		if value, err := value.ToBoolean(); err == nil {
			jsFunctionReturn.Next = value
		} else {
			return errorReturnFilter(err)
		}
	} else {
		return errorReturnFilter(err)
	}

	if value, err := result.Get("body"); err == nil {
		if value, err := js.Call("JSON.stringify", nil, value); err == nil {
			jsFunctionReturn.Body = value.String()
		} else {
			return errorReturnFilter(err)
		}
	} else {
		return errorReturnFilter(err)
	}

	if value, err := result.Get("operation"); err == nil {
		if value, err := value.ToInteger(); err == nil {
			jsFunctionReturn.Operation = model.BodyOperation(value)
		} else {
			return errorReturnFilter(err)
		}
	} else {
		return errorReturnFilter(err)
	}

	if value, err := result.Get("status"); err == nil {
		if value, err := value.ToInteger(); err == nil {
			jsFunctionReturn.Status = int(value)
		} else {
			return errorReturnFilter(err)
		}
	} else {
		return errorReturnFilter(err)
	}

	if value, err := result.Get("error"); err == nil {
		if value.IsDefined() {
			if value, err := value.ToString(); err == nil {
				jsFunctionReturn.Err = errors.New(value)
			} else {
				return errorReturnFilter(err)
			}
		}
	} else {
		return errorReturnFilter(err)
	}
	// weird
	return jsFunctionReturn, jsFunctionReturn.Err
}

func errorReturnFilter(err error) (model.FilterReturn, error) {
	logrus.WithFields(logrus.Fields{
		"error": err.Error(),
	}).Errorf("Error parsing filter return value - js filter exec")
	return model.FilterReturn{Body: "{\"message\" : \"Proxy error : \"" + err.Error() + "}"}, err
}

func httpRequestTOJSContext(call otto.FunctionCall) otto.Value {

	method, err := call.Argument(0).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing httpRequestTOJSContext method - js filter exec")
	}

	url, err := call.Argument(1).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing httpRequestTOJSContext url - js filter exec")
	}

	body, err := call.Argument(2).ToString()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing httpRequestTOJSContext body - js filter exec")
	}

	var req *http.Request

	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Errorf("Error parsing httpRequestTOJSContext GET header - js filter exec")
		}
	} else {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Errorf("Error parsing httpRequestTOJSContext OTHER THAN GET header - js filter exec")
		}
	}

	headers := call.Argument(3).Object()
	if headers != nil {
		for _, key := range headers.Keys() {
			header, err := headers.Get(key)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Errorf("Error parsing httpRequestTOJSContext GET header - js filter exec")
			}
			headerValue, err := header.ToString()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Errorf("Error parsing httpRequestTOJSContext GET header - js filter exec")
			}
			req.Header.Add(key, headerValue)
		}
	}

	logrus.WithFields(logrus.Fields{
		"method": method,
		"url": url,
		"body": body,
		"headers": fmt.Sprintf("%#v", headers),
	}).Infof("Request parameters")

	resp, err := client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error executing the request - httpRequestTOJSContext")
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
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing request body - httpRequestTOJSContext ")
	}

	response, _ := call.Otto.Object("({})")

	bodyObjectJs, err := call.Otto.ToValue(string(bodyBytes))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing response body to JS object - httpRequestTOJSContext")
	}

	headersObjectJs, err := call.Otto.ToValue(resp.Header)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Error parsing response headers to JS object - httpRequestTOJSContext")
	}

	response.Set("body", bodyObjectJs)
	response.Set("status", resp.StatusCode)
	response.Set("headers", headersObjectJs)

	value, _ := otto.ToValue(response)

	return value
}
