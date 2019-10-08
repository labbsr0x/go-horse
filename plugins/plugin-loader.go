package plugins

import (
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/model"
	"github.com/kataras/iris"
	"github.com/robertkrimen/otto"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"plugin"
)

// FilterPluginList filters
var FilterPluginList []GoFilterDefinition

// JSPluginList plugins to set functions in JS context
var JSPluginList []JSContextInjection

// Filter Filter
type GoFilterDefinition interface {
	Config() model.FilterConfig
	Exec(ctx iris.Context, requestBody string) (model.FilterReturn, error)
}

// JSContextInjection JSContextInjection
type JSContextInjection interface {
	Set(ctx iris.Context, call otto.FunctionCall) otto.Value
	Name() string
}

// Load Load
func Load() []GoFilterDefinition {

	if FilterPluginList != nil || JSPluginList != nil {
		return FilterPluginList
	}

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

		var filter GoFilterDefinition
		filter, ok := symPlugin.(GoFilterDefinition)
		if ok {
			FilterPluginList = append(FilterPluginList, filter)
			log.Debug().Str("plugin_name", filter.Config().Name).Str("type", "filter").Msg("Plugin loaded")
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
