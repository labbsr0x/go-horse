package config

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// DockerSockURL DockerSockURL
var DockerSockURL = defaultValue("DOCKER_SOCK", "unix:///var/run/docker.sock")

// TargetHostname TargetHostname
var TargetHostname = defaultValue("TARGET_HOSTNAME", "http://localhost")

// LogLevel LogLevel
var LogLevel = fixLogLevel(defaultValue("LOG_LEVEL", "debug"))

// PrettyLog PrettyLog
var PrettyLog = defaultValueBol("PRETTY_LOG", true)

// Port Port
var Port = fixPortValue(defaultValue("PORT", ":8080"))

// JsFiltersPath JsFiltersPath
var JsFiltersPath = defaultValue("JS_FILTERS_PATH", "/app/go-horse/filters")

// GoPluginsPath GoPluginsPath
var GoPluginsPath = defaultValue("GO_PLUGINS_PATH", "/app/go-horse/plugins")

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
