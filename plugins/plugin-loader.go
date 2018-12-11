package plugins

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"plugin"

	"github.com/kataras/iris"
	"github.com/robertkrimen/otto"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
)

// FilterPluginList lero-lero
var FilterPluginList []Filter

// JSPluginList lero-lero
var JSPluginList []JSContextInjection

// Filter ler-lero
type Filter interface {
	Config() (Name string, Order int, PathPattern string, Invoke int)
	Exec(ctx iris.Context, requestBody string) (Next bool, Body string, Status int, Operation int)
}

// JSContextInjection lero-lero
type JSContextInjection interface {
	Set(ctx iris.Context, call otto.FunctionCall) otto.Value
	Name() string
}

func init() {
	Load()
}

// Load lero-lero
func Load() []Filter {

	FilterPluginList = FilterPluginList[:0]
	JSPluginList = JSPluginList[:0]

	files, err := ioutil.ReadDir(config.GoPluginsPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {

		fmt.Println(">>>>>> PLUGIN : ", file)

		// load module
		// 1. open the so file to load the symbols
		plug, err := plugin.Open(config.GoPluginsPath + "/" + file.Name())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// 2. look up a symbol (an exported function or variable)
		// in this case, variable Plugin
		symPlugin, err := plug.Lookup("Plugin")
		if err != nil {
			fmt.Println("erro ao carregar o plugin : ", err)
			os.Exit(1)
		}

		// 3. Assert that loaded symbol is of a desired type
		// in this case interface type Greeter (defined above)
		var filter Filter
		filter, ok := symPlugin.(Filter)
		if ok {
			FilterPluginList = append(FilterPluginList, filter)
		}

		var js JSContextInjection
		js, ok = symPlugin.(JSContextInjection)
		if ok {
			JSPluginList = append(JSPluginList, js)
		}

	}
	return FilterPluginList

}
