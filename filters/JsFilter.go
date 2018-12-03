package filters

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/security"

	"github.com/kataras/iris"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"github.com/robertkrimen/otto"
)

var jsFilterFunctions []string

// FilterMoldels lero lero
var FilterMoldels []JsFilterModel

// FiltersBefore lero lero
var FiltersBefore []JsFilterModel

// FiltersAfter lero lero
var FiltersAfter []JsFilterModel

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
	js.Set("method", strings.ToUpper(ctx.Method()))

	returnValue, error := js.Run("(" + model.Function + ")(url, body, operation, method, verifyPolicy)")

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

func init() {
	readFromFile()
	parseFilterObject()
}

func readFromFile() {

	files, err := ioutil.ReadDir(config.JsFiltersPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		content, error := ioutil.ReadFile(config.JsFiltersPath + "/" + file.Name())
		if error != nil {
			continue
		}
		jsFilterFunctions = append(jsFilterFunctions, string(content))
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

	for _, jsFunc := range jsFilterFunctions {
		js := otto.New()

		invokeObj, error := js.Object("({AFTER : 0, BEFORE : 1})")
		if error != nil {
			fmt.Println("ERRO AO CRIAR O JS OBJECT 'Invoke'.", error)
		}
		js.Set("invoke", invokeObj)

		funcFilterDefinition, error := js.Call("(function(invoke){return"+jsFunc+"})", nil, invokeObj)
		if error != nil {
			fmt.Println("ERRO AO TENTAR PARSEAR A DEFINICAO DO FILTRO ::::: ", error, "\n", jsFunc)
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
			FiltersBefore = append(FiltersBefore, filterDefinition)
		} else {
			FiltersAfter = append(FiltersAfter, filterDefinition)
		}

		orderFilterModels(FilterMoldels, FiltersBefore, FiltersAfter)
		validateFilterOrder(FilterMoldels)

	}

}
