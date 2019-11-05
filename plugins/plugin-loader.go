package plugins

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"plugin"

	"github.com/labbsr0x/go-horse/filters/model"
	"github.com/kataras/iris"
	"github.com/robertkrimen/otto"
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
func Load(goPluginsPath string) []GoFilterDefinition {

	if FilterPluginList != nil || JSPluginList != nil {
		return FilterPluginList
	}

	files, err := ioutil.ReadDir(goPluginsPath)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Errorf("Could not load plugins from directory")
	}

	for _, file := range files {

		logrus.WithFields(logrus.Fields{
			"file": file.Name(),
		}).Debugf("Loading plugin")

		plug, err := plugin.Open(goPluginsPath + "/" + file.Name())
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"plugin_path": goPluginsPath+"/"+file.Name(),
			}).Errorf("Could not open plugin")
		}

		symPlugin, err := plug.Lookup("Plugin")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
				"plugin_path": goPluginsPath+"/"+file.Name(),
			}).Errorf("Could not load plugin")
		}

		var filter GoFilterDefinition
		filter, ok := symPlugin.(GoFilterDefinition)
		if ok {
			FilterPluginList = append(FilterPluginList, filter)
			logrus.WithFields(logrus.Fields{
				"plugin_name": filter.Config().Name,
				"type": "filter",
			}).Debugf("Plugin loaded")
		}

		var js JSContextInjection
		js, ok = symPlugin.(JSContextInjection)
		if ok {
			JSPluginList = append(JSPluginList, js)
			logrus.WithFields(logrus.Fields{
				"plugin_name": js.Name(),
				"type": "js",
			}).Debugf("Plugin loaded")
		}

	}
	return FilterPluginList

}
