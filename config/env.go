package config

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// DockerSockURL lero-lero
var DockerSockURL = defaultValue("DOCKER_SOCK", "unix:///var/run/docker.sock")

// DockerSockURL lero-lero
var TargetHostname = defaultValue("TARGET_HOSTNAME", "http://localhost")

// LogLevel lero lero
var LogLevel = fixLogLevel(defaultValue("LOG_LEVEL", "debug"))

// PrettyLog ler-lero
var PrettyLog = defaultValueBol("PRETTY_LOG", true)

// Port lero lero
var Port = fixPortValue(defaultValue("PORT", ":8080"))

// JsFiltersPath lero lero
var JsFiltersPath = defaultValue("JS_FILTERS_PATH", os.Getenv("HOME")+"/sadman-acl-proxy/filters")

// GoPluginsPath lero-lero
var GoPluginsPath = defaultValue("GO_PLUGINS_PATH", os.Getenv("HOME")+"/sadman-acl-proxy/plugins")

func fixLogLevel(logLevel string) zerolog.Level {
	level, error := zerolog.ParseLevel(logLevel)
	if error != nil {
		return zerolog.WarnLevel
	}
	return level
}

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

func defaultValueBol(envVar string, value bool) bool {
	env := os.Getenv(envVar)
	if env == "" {
		return value
	}
	return env == "true"
}
