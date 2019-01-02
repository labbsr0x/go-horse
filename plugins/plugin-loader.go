package plugins

import (
	"github.com/kataras/iris"
	"github.com/robertkrimen/otto"
)

// FilterPluginList filters
var FilterPluginList []Filter

// JSPluginList plugins to set functions in JS context
var JSPluginList []JSContextInjection

// Filter Filter
type Filter interface {
	Config() (Name string, Order int, PathPattern string, Invoke int)
	Exec(ctx iris.Context, requestBody string) (Next bool, Body string, Status int, Operation int, Err error)
}

// JSContextInjection JSContextInjection
type JSContextInjection interface {
	Set(ctx iris.Context, call otto.FunctionCall) otto.Value
	Name() string
}

// Load Load
func Load() []Filter {

	// if FilterPluginList != nil || JSPluginList != nil {
	// 	return FilterPluginList
	// }

	// files, err := ioutil.ReadDir(config.GoPluginsPath)
	// if err != nil {
	// 	log.Error().Err(err).Str("dir", config.GoPluginsPath).Msg("Could not load plugins from directory")
	// }

	// for _, file := range files {

	// 	log.Debug().Str("file", file.Name()).Msg("Loading plugin")

	// 	plug, err := plugin.Open(config.GoPluginsPath + "/" + file.Name())
	// 	if err != nil {
	// 		log.Error().Err(err).Str("plugin_path", config.GoPluginsPath+"/"+file.Name()).Msg("Could not open plugin")
	// 	}

	// 	symPlugin, err := plug.Lookup("Plugin")
	// 	if err != nil {
	// 		log.Error().Err(err).Str("plugin_path", config.GoPluginsPath+"/"+file.Name()).Msg("Could not load plugin")
	// 	}

	// 	var filter Filter
	// 	filter, ok := symPlugin.(Filter)
	// 	if ok {
	// 		FilterPluginList = append(FilterPluginList, filter)
	// 		name, _, _, _ := filter.Config()
	// 		log.Debug().Str("plugin_name", name).Str("type", "filter").Msg("Plugin loaded")
	// 	}

	// 	var js JSContextInjection
	// 	js, ok = symPlugin.(JSContextInjection)
	// 	if ok {
	// 		JSPluginList = append(JSPluginList, js)
	// 		log.Debug().Str("plugin_name", js.Name()).Str("type", "js").Msg("Plugin loaded")
	// 	}

	// }
	// fmt.Printf("%#v\n", FilterPluginList)
	return FilterPluginList

}
