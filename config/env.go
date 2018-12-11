package config

import (
	"os"
	"strings"
)

// Port lero lero
var Port = fixPortValue(defaultValue("PORT", ":8080"))

// JsFiltersPath lero lero
var JsFiltersPath = defaultValue("JS_FILTERS_PATH", os.Getenv("HOME")+"/sadman-acl-proxy/filters")

// GoPluginsPath lero-lero
var GoPluginsPath = defaultValue("GO_PLUGINS_PATH", os.Getenv("HOME")+"/sadman-acl-proxy/plugins")

func fixPortValue(port string) string {
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func defaultValue(envVar, value string) string {
	env := os.Getenv(envVar)
	if env == "" {
		return value
	}
	return env
}
