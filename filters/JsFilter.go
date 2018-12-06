package filters

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/util"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/security"

	"github.com/kataras/iris"
	"github.com/radovskyb/watcher"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"github.com/robertkrimen/otto"
)

var client = &http.Client{}

var updateLock = sync.WaitGroup{}
var isUpdating = false

var jsFilterFunctions = make(map[string]string)

// FilterMoldels lero lero
var FilterMoldels []JsFilterModel

// FiltersBefore lero lero
var filtersBefore []JsFilterModel

// FiltersAfter lero lero
var filtersAfter []JsFilterModel

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
	Order       int64
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
}

// BeforeFilters lero lero
func BeforeFilters() []JsFilterModel {
	if isUpdating {
		updateLock.Wait()
	}
	return filtersBefore
}

// AfterFilters lero lero
func AfterFilters() []JsFilterModel {
	if isUpdating {
		updateLock.Wait()
	}
	return filtersAfter
}

// MatchURL lero lero
func (model JsFilterModel) MatchURL(ctx iris.Context) bool {
	if model.regex == nil {
		regex, error := regexp.Compile(model.PathPattern)
		if error != nil {
			fmt.Printf("ERRO AO CRIAR REGEX PARA DAR MATCH NA URL DO FILTRO : %s; PATTERN : %s\n", model.Name, model.PathPattern)
		} else {
			model.regex = regex
		}
	}
	return model.regex.MatchString(ctx.RequestPath(false))
}

// ExecResponse lerol ero
func (model JsFilterModel) ExecResponse(ctx iris.Context, response *http.Response) JsFilterFunctionReturn {
	body, erro := ioutil.ReadAll(response.Body)
	if erro != nil {
		fmt.Println("Erro parsear body para execução da função javascript ::>>" + erro.Error())
	}
	return model.Exec(ctx, string(body))
}

// Exec lero lero
func (model JsFilterModel) Exec(ctx iris.Context, body string) JsFilterFunctionReturn {

	js := otto.New()
	js.Set("url", ctx.Request().URL.Path)

	funcRet, error := js.Call("JSON.parse", nil, body)
	if error != nil {
		fmt.Println("ERRO AO RODAR O JSON>PARSE ::::: ", error)
		emptyBody, _ := js.Object("({})")
		funcRet, _ = otto.ToValue(emptyBody)
	}

	js.Set("body", funcRet.Object())

	operation, error := js.Object("({READ : 0, WRITE : 1})")
	if error != nil {
		fmt.Println("ERRO AO CRIAR O JS OBJECT 'Operation'.", error)
	}
	js.Set("operation", operation)
	js.Set("verifyPolicy", veryfyPolicyToJSContext)
	js.Set("getVar", func(call otto.FunctionCall) otto.Value { return requestScopeGetToJSContext(ctx, call) })
	js.Set("setVar", func(call otto.FunctionCall) otto.Value { return requestScopeSetToJSContext(ctx, call) })
	js.Set("listVar", func(call otto.FunctionCall) otto.Value { return requestScopeListToJSContext(ctx, call) })
	js.Set("method", strings.ToUpper(ctx.Method()))
	js.Set("headers", ctx.Request().Header)
	js.Set("request", httpRequestTOJSContext)

	returnValue, error := js.Run("(" + model.Function + ")(url, body, operation, method, verifyPolicy, getVar, setVar, listVar, headers, request)")

	if error != nil {
		fmt.Println("Erro executar fn js ::>> ", error)
	}

	result := returnValue.Object()

	if error != nil {
		fmt.Println("Erro executar ao tentar obter o valor de retorno da função js ::>> ", error)
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
	return JsFilterFunctionReturn{Next: false, Body: "{\"message\" : \"Erro filtro sandman acl : \"" + error.Error() + "}"}
}

func veryfyPolicyToJSContext(call otto.FunctionCall) otto.Value { //
	method, error := call.Argument(0).ToString()
	if error != nil {
		fmt.Println("ERRO : parametro method : ", error)
	}
	url, error := call.Argument(1).ToString()
	if error != nil {
		fmt.Println("ERRO : parametro url : ", error)
	}
	allowed := security.VerifyPolicy(method, url)
	result, error := otto.ToValue(allowed)
	if error != nil {
		fmt.Println("ERRO : returno da função javascript  : ", error)
	}
	return result
}

func requestScopeGetToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value { //
	key, error := call.Argument(0).ToString()
	if error != nil {
		fmt.Println("ERRO : parametro key : ", error)
	}
	value := util.RequestScopeGet(ctx, key)
	result, error := otto.ToValue(value)
	if error != nil {
		fmt.Println("ERRO : returno da função javascript  : ", error)
	}
	return result
}

func requestScopeSetToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value { //
	key, error := call.Argument(0).ToString()
	if error != nil {
		fmt.Println("ERRO : parametro key : ", error)
	}
	value, error := call.Argument(1).ToString()
	if error != nil {
		fmt.Println("ERRO : parametro value : ", error)
	}
	util.RequestScopeSet(ctx, key, value)
	return otto.NullValue()
}

func requestScopeListToJSContext(ctx iris.Context, call otto.FunctionCall) otto.Value { //
	mapa := util.RequestScopeList(ctx)
	result, error := call.Otto.ToValue(mapa)
	if error != nil {
		fmt.Println("ERRO : tentando transformar o mapa do retorno da func RequestScopeList to JS value : ", error)
	}
	return result
}

func httpRequestTOJSContext(call otto.FunctionCall) otto.Value {
	method, error := call.Argument(0).ToString()
	if error != nil {
		fmt.Println("Erro ao parsear o parametro method do request js->go : ", error)
	}
	url, error := call.Argument(1).ToString()
	if error != nil {
		fmt.Println("Erro ao parsear o parametro url do request js->go : ", error)
	}
	body, error := call.Argument(3).ToString()
	if error != nil {
		fmt.Println("Erro ao parsear o parametro body do request js->go : ", error)
	}
	var req *http.Request
	var err interface{}

	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			fmt.Println("Erro ao parsear o parametro header do request js->go : ", error)
		}
	} else {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
		if err != nil {
			fmt.Println("Erro ao parsear o parametro header do request js->go : ", error)
		}
	}

	headers := call.Argument(4).Object()
	if headers != nil {
		for _, key := range headers.Keys() {
			header, error := headers.Get(key)
			if error != nil {
				fmt.Println("Erro ao parsear o parametro header do request js->go : ", error)
			}
			headerValue, error := header.ToString()
			if error != nil {
				fmt.Println("Erro ao parsear o parametro header do request js->go : ", error)
			}
			req.Header.Add(key, headerValue)
		}
	}
	fmt.Println("************************** PARAMETROS REQUEST : %s, %s, %s, %+v", method, url, body, headers)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Erro ao executar o request : ", err)
	}
	defer resp.Body.Close()
	if req.Body != nil {
		defer req.Body.Close()
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao parsear resposta do request request : ", err)
	}

	result, error := call.Otto.ToValue(string(bodyBytes))
	if error != nil {
		fmt.Println("Erro ao transformar o retorno do request js->go : ", error)
	}

	return result
}

func init() {
	readFromFile()
	parseFilterObject()
	dirWatcher()
}

func updateFilters() {
	updateLock.Add(1)

	isUpdating = true

	readFromFile()
	parseFilterObject()
	updateLock.Done()

	isUpdating = false

}

func dirWatcher() {
	dirWatcher := watcher.New()

	go func() {
		for {
			select {
			case event := <-dirWatcher.Event:
				fmt.Println(event) // Print the event's info.
				updateFilters()
			case err := <-dirWatcher.Error:
				log.Fatalln("\n\n########### " + err.Error() + " ###########\n\n")
			case <-dirWatcher.Closed:
				return
			}
		}
	}()

	// // Watch this folder for changes.
	// if err := dirWatcher.AddRecursive("../"); err != nil {
	// 	log.Fatalln(err)
	// }

	if err := dirWatcher.AddRecursive(config.JsFiltersPath); err != nil {
		log.Fatalln(err)
	}

	// go func() {
	// 	dirWatcher.Wait()
	// 	dirWatcher.TriggerEvent(watcher.Create, nil)
	// 	dirWatcher.TriggerEvent(watcher.Remove, nil)
	// }()

	go func() {
		if err := dirWatcher.Start(time.Second); err != nil {
			log.Fatalln(err)
		}
	}()
}

func readFromFile() {

	jsFilterFunctions = make(map[string]string)

	files, err := ioutil.ReadDir(config.JsFiltersPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		content, error := ioutil.ReadFile(config.JsFiltersPath + "/" + file.Name())
		if error != nil {
			continue
		}
		jsFilterFunctions[file.Name()] = string(content)
		fmt.Println(file.Name(), " >> ", string(content))
	}

}

func orderFilterModels(models ...[]JsFilterModel) {
	for _, filters := range models {
		sort.SliceStable(filters[:], func(i, j int) bool {
			return filters[i].Order < filters[j].Order
		})
	}
}

func validateFilterOrder(models []JsFilterModel) {
	var last int64 = -1
	for _, filter := range models {
		fmt.Println("order : ", filter.Order)
		fmt.Println("last : ", last)
		if filter.Order == last {
			panic(fmt.Sprintf("Erro na definição dos filtros : colisão da propriedade ordem : existem 2 filtros com a ordem nro -> %d", last))
		}
		last = filter.Order
	}
}

func parseFilterObject() {

	FilterMoldels = FilterMoldels[:0]
	filtersBefore = filtersBefore[:0]
	filtersAfter = filtersAfter[:0]

	for fileName, jsFunc := range jsFilterFunctions {
		js := otto.New()

		invokeObj, error := js.Object("({AFTER : 0, BEFORE : 1})")
		if error != nil {
			fmt.Println("ERRO AO CRIAR O JS OBJECT 'Invoke'.", error)
		}
		js.Set("invoke", invokeObj)

		funcFilterDefinition, error := js.Call("(function(invoke){return"+jsFunc+"})", nil, invokeObj)
		if error != nil {
			fmt.Println("ERRO AO TENTAR PARSEAR A DEFINICAO DO FILTRO ::::: ", error, "\n", jsFunc)
			fmt.Println(">>>>>>>>>>>>>>> IGNORANDO O ARQUIVO ", fileName)
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
				//(error)
			}
		} else {
			//(error)
		}

		if value, err := filter.Get("name"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.Name = value
			} else {
				//(error)
			}
		} else {
			//(error)
		}

		if value, err := filter.Get("order"); err == nil {
			if value, err := value.ToInteger(); err == nil {
				filterDefinition.Order = value
			} else {
				//(error)
			}
		} else {
			//(error)
		}

		if value, err := filter.Get("pathPattern"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.PathPattern = value
			} else {
				//(error)
			}
		} else {
			//(error)
		}

		if value, err := filter.Get("function"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.Function = value
			} else {
				//(error)
			}
		} else {
			//(error)
		}

		FilterMoldels = append(FilterMoldels, filterDefinition)

		if filterDefinition.Invoke == Before {
			filtersBefore = append(filtersBefore, filterDefinition)
		} else {
			filtersAfter = append(filtersAfter, filterDefinition)
		}

		orderFilterModels(FilterMoldels, filtersBefore, filtersAfter)
		validateFilterOrder(FilterMoldels)

	}

}
