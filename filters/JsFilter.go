package filters

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/kataras/iris"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"github.com/robertkrimen/otto"
)

var jsFilterFunctions []string

// FilterMoldels lero lero
var FilterMoldels []JsFilterModel

// JsFilterModel lero-lero
type JsFilterModel struct {
	Name        string
	Order       int64
	PathPattern string
	Function    string
}

// ExecResponse lerol ero
func (model JsFilterModel) ExecResponse(ctx iris.Context, response *http.Response) bool {
	body, erro := ioutil.ReadAll(response.Body)
	if erro != nil {
		fmt.Println("Erro parsear body para execução da função javascript ::>>" + erro.Error())
	}
	return model.Exec(ctx, string(body))
}

// Exec lero lero
func (model JsFilterModel) Exec(ctx iris.Context, body string) bool {
	js := otto.New()
	js.Set("url", ctx.Request().URL.Path)

	funcRet, error := js.Call("JSON.parse", nil, body)
	if error != nil {
		fmt.Println("ERRO AO RODAR O JSON>PARSE ::::: ", error)
	}

	js.Set("body", funcRet.Object())

	returnValue, error := js.Run("(" + model.Function + ")(url, body)")

	ret, error := returnValue.ToBoolean()

	if error != nil {
		fmt.Println("Erro executar fn js ::>> ", error)
		return false
	}

	return ret
}

func init() {
	readFromFile()
	parseFilterObject()
	fmt.Println(FilterMoldels)
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

	fmt.Println(jsFilterFunctions)

}

func parseFilterObject() {

	for _, jsFunc := range jsFilterFunctions {
		js := otto.New()

		// funcRet, error := js.Call("JSON.parse", nil, "{\"teste\": 123}")
		// if error != nil {
		// 	fmt.Println("ERRO AO RODAR O JSON>PARSE ::::: ", error)
		// }
		// val, _ := funcRet.Object().Get("teste")

		// fmt.Println("JSON>PARSE>TESTE>VALUE>123 :: ", val.String())

		filter, error := js.Object("(" + jsFunc + ")")
		if error != nil {
			fmt.Println(error)
		}

		filterDefinition := JsFilterModel{}

		if value, err := filter.Get("name"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.Name = value
			}
		}

		if value, err := filter.Get("order"); err == nil {
			if value, err := value.ToInteger(); err == nil {
				filterDefinition.Order = value
			}
		}

		if value, err := filter.Get("pathPattern"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.PathPattern = value
			}
		}

		if value, err := filter.Get("function"); err == nil {
			if value, err := value.ToString(); err == nil {
				filterDefinition.Function = value
			}
		}

		FilterMoldels = append(FilterMoldels, filterDefinition)

	}

}
