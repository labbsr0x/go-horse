package plugins

import (
	"io/ioutil"
	"plugin"

	"github.com/kataras/iris"
	"github.com/robertkrimen/otto"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"github.com/rs/zerolog/log"
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
		log.Error().Err(err).Str("dir", config.GoPluginsPath).Msg("Could not load plugins from directory")
	}

	for _, file := range files {

		log.Debug().Str("file", file.Name()).Msg("Loading plugin")

		plug, err := plugin.Open(config.GoPluginsPath + "/" + file.Name())
		if err != nil {
			log.Error().Err(err).Str("plugin_path", config.GoPluginsPath+"/"+file.Name()).Msg("Could not open plugin")
		}

		symPlugin, err := plug.Lookup("Plugin")
		if err != nil {
			log.Error().Err(err).Str("plugin_path", config.GoPluginsPath+"/"+file.Name()).Msg("Could not load plugin")
		}

		var filter Filter
		filter, ok := symPlugin.(Filter)
		if ok {
			FilterPluginList = append(FilterPluginList, filter)
			name, _, _, _ := filter.Config()
			log.Debug().Str("plugin_name", name).Str("type", "filter").Msg("Plugin loaded")
		}

		var js JSContextInjection
		js, ok = symPlugin.(JSContextInjection)
		if ok {
			JSPluginList = append(JSPluginList, js)
			log.Debug().Str("plugin_name", js.Name()).Str("type", "js").Msg("Plugin loaded")
		}

	}
	return FilterPluginList

}
