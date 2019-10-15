package web

import (
	"os"
)

// // DockerAPIVersion DockerAPIVersion
// var DockerAPIVersion = defaultValue("DOCKER_API_VERSION", "1.39")

// // DockerSockURL DockerSockURL
// var DockerSockURL = defaultValue("DOCKER_SOCK", "unix:///var/run/docker.sock")

// // TargetHostname TargetHostname
// var TargetHostname = defaultValue("TARGET_HOSTNAME", "http://go-horse")

// // LogLevel LogLevel
// var LogLevel = fixLogLevel(defaultValue("LOG_LEVEL", "info"))

// // PrettyLog PrettyLog
// var PrettyLog = defaultValueBol("PRETTY_LOG", true)

// // Port Port
// var Port = fixPortValue(defaultValue("PORT", ":8080"))

// JsFiltersPath JsFiltersPath
var JsFiltersPath = defaultValue("JS_FILTERS_PATH", "/app/go-horse/filters")

// GoPluginsPath GoPluginsPath
var GoPluginsPath = defaultValue("GO_PLUGINS_PATH", "/app/go-horse/plugins")

// // SetJsFiltersPath SetJsFiltersPath for e2e tests
// func SetJsFiltersPath(path string) {
// 	JsFiltersPath = path
// }

// // SetGoPluginsPath SetGoPluginsPath for e2e tests
// func SetGoPluginsPath(path string) {
// 	GoPluginsPath = path
// }

// // SetPort helper function to change port for e2e tests
// func SetPort(port string) {
// 	Port = fixPortValue(port)
// }

// // SetLogLevel helper function to change log level for e2e tests
// func SetLogLevel(level string) {
// 	LogLevel = fixLogLevel(level)
// }

// func fixLogLevel(logLevel string) zerolog.Level {
// 	level, error := zerolog.ParseLevel(logLevel)
// 	if error != nil {
// 		return zerolog.WarnLevel
// 	}
// 	return level
// }

// func fixPortValue(port string) string {
// 	if strings.HasPrefix(port, ":") {
// 		return port
// 	}
// 	return ":" + port
// }

func defaultValue(envVar, value string) string {
	env := os.Getenv(envVar)
	if env == "" {
		return value
	}
	return env
}

// func defaultValueBol(envVar string, value bool) bool {
// 	env := os.Getenv(envVar)
// 	if env == "" {
// 		return value
// 	}
// 	return env == "true"
// }
