package filters

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/plugins"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/util"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/security"

	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"github.com/robertkrimen/otto"
)

var client = &http.Client{}

// BodyOperation json body operation type
type BodyOperation int

const (
	// Read read
	Read BodyOperation = 0
	// Write write
	Write BodyOperation = 1
)

// Invoke invoke
type Invoke int

const (
	// After After
	After Invoke = 0
	// Before After
	Before Invoke = 1
)

// JsFilterModel lero-lero
type JsFilterModel struct {
	Name        string
	Order       int
	PathPattern string
	Invoke      Invoke
	Function    string
	regex       *regexp.Regexp
}

// JsFilterFunctionReturn lero-lero
type JsFilterFunctionReturn struct {
	Next      bool
	Body      string
	Status    int
	Operation BodyOperation
	Err       error
}

// MatchURL lero lero
func (jsFilter JsFilterModel) MatchURL(ctx iris.Context) bool {
	if jsFilter.regex == nil {
		regex, error := regexp.Compile(jsFilter.PathPattern)
		if error != nil {
			log.Error().Str("plugin_name", jsFilter.Name).Err(error).Msg("Error compiling the filter url matcher regex")
		} else {
			jsFilter.regex = regex
		}
	}
	return jsFilter.regex.MatchString(ctx.RequestPath(false))
}

// ExecResponse lerol ero
func (jsFilter JsFilterModel) ExecResponse(ctx iris.Context, response *http.Response) JsFilterFunctionReturn {
	body, erro := ioutil.ReadAll(response.Body)
	if erro != nil {
		log.Error().Str("plugin_name", jsFilter.Name).Err(erro).Msg("Error parsing body")
	}
	return jsFilter.Exec(ctx, string(body))
}

// Exec lero lero
func (jsFilter JsFilterModel) Exec(ctx iris.Context, body string) JsFilterFunctionReturn {

	js := otto.New()
	js.Set("url", ctx.Request().URL.Path)

	funcRet, error := js.Call("JSON.parse", nil, body)
	if error != nil {
		log.Error().Str("plugin_name", jsFilter.Name).Err(error).Msg("Error parsing body string to JS object - js filter exec")
		emptyBody, _ := js.Object("({})")
		funcRet, _ = otto.ToValue(emptyBody)
	}

	js.Set("body", funcRet.Object())

	operation, error := js.Object("({READ : 0, WRITE : 1})")
	if error != nil {
		log.Error().Str("plugin_name", jsFilter.Name).Err(error).Msg("Error creating operation object - js filter exec")
	}
	js.Set("operation", operation)
	js.Set("verifyPolicy", veryfyPolicyToJSContext)
	js.Set("getVar", func(call otto.FunctionCall) otto.Value { return requestScopeGetToJSContext(ctx, call) })
	js.Set("setVar", func(call otto.FunctionCall) otto.Value { return requestScopeSetToJSContext(ctx, call) })
	js.Set("listVar", func(call otto.FunctionCall) otto.Value { return requestScopeListToJSContext(ctx, call) })
	js.Set("method", strings.ToUpper(ctx.Method()))
	js.Set("headers", ctx.Request().Header)
	js.Set("request", httpRequestTOJSContext)

	pluginsJsObj, error := js.Object("({})")

	for _, jsPlugin := range plugins.JSPluginList {
		error := pluginsJsObj.Set(jsPlugin.Name(), func(call otto.FunctionCall) otto.Value { return jsPlugin.Set(ctx, call) })
		if error != nil {
			log.Error().Str("plugin_name", jsPlugin.Name()).Err(error).Msg("Error on applying GO->JS plugin - js filter exec")
		}
	}

	js.Set("plugins", pluginsJsObj)

	returnValue, error := js.Run("(" + jsFilter.Function + ")(url, body, operation, method, verifyPolicy, getVar, setVar, listVar, headers, request, plugins)")

	if error != nil {
		log.Error().Str("plugin_name", jsFilter.Name).Err(error).Msg("Error executing filter - js filter exec")
	}

	result := returnValue.Object()

	if error != nil {
		log.Error().Str("plugin_name", jsFilter.Name).Err(error).Msg("Error parsing return value from filter - js filter exec")
		return JsFilterFunctionReturn{Next: false, Body: "{\"message\" : \"Erro filtro sandman acl : \"" + error.Error() + "}"}
	}

	jsFunctionReturn := JsFilterFunctionReturn{}

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
			if value == 1 {
				jsFunctionReturn.Operation = Write
			} else {
				jsFunctionReturn.Operation = Read
			}
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

	return jsFunctionReturn
}

func errorReturnFilter(error error) JsFilterFunctionReturn {
	log.Error().Err(error).Msg("Error parsing filter return value - js filter exec")
	return JsFilterFunctionReturn{Next: false, Body: "{\"message\" : \"Proxy error : \"" + error.Error() + "}"}
}

func veryfyPolicyToJSContext(call otto.FunctionCall) otto.Value { //
	method, error := call.Argument(0).ToString()
	if error != nil {
		log.Error().Err(error).Msg("Error parsing veryfyPolicyToJSContext method field - js filter exec")
	}
	url, error := call.Argument(1).ToString()
	if error != nil {
		log.Error().Err(error).Msg("Error parsing veryfyPolicyToJSContext url field - js filter exec")
	}
	allowed := security.VerifyPolicy(method, url)
	result, error := otto.ToValue(allowed)
	if error != nil {
		log.Error().Err(error).Msg("Error parsing veryfyPolicyToJSContext function field - js filter exec")
	}
	return result
}

func requestScopeGetToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value { //
	key, error := call.Argument(0).ToString()
	if error != nil {
		log.Error().Err(error).Msg("Error parsing requestScopeGetToJSContext key field - js filter exec")
	}
	value := util.RequestScopeGet(ctx, key)
	result, error := otto.ToValue(value)
	if error != nil {
		log.Error().Err(error).Msg("Error parsing requestScopeGetToJSContext function return - js filter exec")
	}
	return result
}

func requestScopeSetToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value { //
	key, error := call.Argument(0).ToString()
	if error != nil {
		log.Error().Err(error).Msg("Error parsing requestScopeSetToJSContext key field - js filter exec")
	}
	value, error := call.Argument(1).ToString()
	if error != nil {
		log.Error().Err(error).Msg("Error parsing requestScopeSetToJSContext function exec - js filter exec")
	}
	util.RequestScopeSet(ctx, key, value)
	return otto.NullValue()
}

func requestScopeListToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value { //
	mapa := util.RequestScopeList(ctx)
	result, error := call.Otto.ToValue(mapa)
	if error != nil {
		log.Error().Err(error).Msg("Error parsing requestScopeListToJSContext response map - js filter exec")
	}
	return result
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
	}
	defer resp.Body.Close()
	if req.Body != nil {
		defer req.Body.Close()
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Msg("Error parsing request body - httpRequestTOJSContext " + fmt.Sprintf("%#v", err))
	}

	response, error := call.Otto.Object("({})")
	if error != nil {
		log.Error().Err(error).Msg("Error creating response js object - httpRequestTOJSContext")
	}

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

// Load lero-lero
func Load() []JsFilterModel {
	return parseFilterObject(readFromFile())
}

func readFromFile() map[string]string {

	var jsFilterFunctions = make(map[string]string)

	files, err := ioutil.ReadDir(config.JsFiltersPath)
	if err != nil {
		log.Error().Err(err).Msg("Error reading filters dir - readFromFile")
	}

	for _, file := range files {
		content, error := ioutil.ReadFile(config.JsFiltersPath + "/" + file.Name())
		if error != nil {
			log.Error().Err(err).Str("file", file.Name()).Msg("Error reading filter filter - readFromFile")
			continue
		}
		jsFilterFunctions[file.Name()] = string(content)
		log.Debug().Str("file", file.Name()).Str("filter_content", string(content)).Msg("js filter - readFromFile")
	}

	return jsFilterFunctions
}

func parseFilterObject(jsFilterFunctions map[string]string) []JsFilterModel {

	var filterMoldels []JsFilterModel

	for fileName, jsFunc := range jsFilterFunctions {
		js := otto.New()

		invokeObj, error := js.Object("({AFTER : 0, BEFORE : 1})")
		if error != nil {
			log.Error().Err(error).Str("file", fileName).Msg("Error creating invoke object - parseFilterObject")
		}
		js.Set("invoke", invokeObj)

		funcFilterDefinition, error := js.Call("(function(invoke){return"+jsFunc+"})", nil, invokeObj)
		if error != nil {
			log.Error().Err(error).Str("file", fileName).Msg("Error on JS object definition - parseFilterObject")
			continue
		}

		filter := funcFilterDefinition.Object()

		filterDefinition := JsFilterModel{}

		if value, err := filter.Get("invoke"); err == nil {
			if value, err := value.ToInteger(); err == nil {
				if value == 1 {
					filterDefinition.Invoke = Before
				} else {
					filterDefinition.Invoke = After
				}
			} else {
				log.Error().Err(err).Str("file", fileName).Str("field", "invoke").Msg("Error on JS filter definition - parseFilterObject")
			}
		} else {
			log.Error().Err(err).Str("file", fileName).Str("field", "invoke").Msg("Error on JS filter definition - parseFilterObject")
		}

		if value, err := filter.Get("name"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.Name = value
			} else {
				log.Error().Err(err).Str("file", fileName).Str("field", "name").Msg("Error on JS filter definition - parseFilterObject")
			}
		} else {
			log.Error().Err(err).Str("file", fileName).Str("field", "name").Msg("Error on JS filter definition - parseFilterObject")
		}

		if value, err := filter.Get("order"); err == nil {
			if value, err := value.ToInteger(); err == nil {
				filterDefinition.Order = int(value)
			} else {
				log.Error().Err(err).Str("file", fileName).Str("field", "order").Msg("Error on JS filter definition - parseFilterObject")
			}
		} else {
			log.Error().Err(err).Str("file", fileName).Str("field", "order").Msg("Error on JS filter definition - parseFilterObject")
		}

		if value, err := filter.Get("pathPattern"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.PathPattern = value
			} else {
				log.Error().Err(err).Str("file", fileName).Str("field", "pathPattern").Msg("Error on JS filter definition - parseFilterObject")
			}
		} else {
			log.Error().Err(err).Str("file", fileName).Str("field", "pathPattern").Msg("Error on JS filter definition - parseFilterObject")
		}

		if value, err := filter.Get("function"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.Function = value
			} else {
				log.Error().Err(err).Str("file", fileName).Str("field", "function").Msg("Error on JS filter definition - parseFilterObject")
			}
		} else {
			log.Error().Err(err).Str("file", fileName).Str("field", "function").Msg("Error on JS filter definition - parseFilterObject")
		}

		filterMoldels = append(filterMoldels, filterDefinition)
	}
	return filterMoldels
}
