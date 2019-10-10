package config

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// All envs that GHP need to work with
const (
	dockerAPIVersion = "docker-api-version"
	dockerSockURL    = "docker-sock-url"
	targetHostName   = "target-host-name"
	logLevel         = "log-level"
	prettyLog        = "pretty-log"
	port             = "port"
	jsFiltersPath    = "js-filter-path"
	goPluginsPath    = "go-plugins-path"
)

// Flags define the fields that will be passed via cmd
type Flags struct {
	DockerAPIVersion string
	DockerSockURL    string
	TargetHostName   string
	LogLevel         string
	PrettyLog        bool // Bool or string ?
	Port             string
	JsFiltersPath    string
	GoPluginsPath    string
}

// WebBuilder defines the parametric information of a gohorse server instance
type WebBuilder struct {
	*Flags
}

// AddFlags adds flags for Builder.
// TODO : Discuss shortchut name
func AddFlags(flags *pflag.FlagSet) {
	flags.StringP(dockerAPIVersion, "v", "1.39", "Version of Docker API")
	flags.StringP(dockerSockURL, "u", "unix:///var/run/docker.sock", "URL of Docker Socket")
	flags.StringP(targetHostName, "n", "http://go-horse", "Target host name")
	flags.StringP(logLevel, "l", "info", "[optional] Sets the Log Level to one of seven (trace, debug, info, warn, error, fatal, panic). Defaults to info")
	flags.BoolP(prettyLog, "t", false, "Enable or disable pretty log. Defaults to false")
	flags.StringP(port, "p", ":8080", "Go Horse port. Defaults to :8080")
	flags.StringP(jsFiltersPath, "j", "/app/go-horse/filters", "Sets the path to json filters")
	flags.StringP(goPluginsPath, "g", "/app/go-horse/plugins", "Sets the path to go plugins")
}

// InitFromWebBuilder initializes the web server builder with properties retrieved from Viper.
func (b *WebBuilder) InitFromViper(v *viper.Viper) *WebBuilder {
	flags := new(Flags)
	flags.DockerAPIVersion = v.GetString(dockerAPIVersion)
	flags.TargetHostName = v.GetString(targetHostName)
	flags.LogLevel = v.GetString(logLevel)
	flags.PrettyLog = v.GetBool(prettyLog)
	flags.Port = v.GetString(port)
	flags.JsFiltersPath = v.GetString(jsFiltersPath)
	flags.GoPluginsPath = v.GetString(goPluginsPath)

	flags.check()

	b.Flags = flags

	return b
}

func (flags *Flags) check() {

	logrus.Infof("Flags: '%v'", flags)

	haveEmptyRequiredFlags := flags.DockerAPIVersion == "" ||
		flags.TargetHostName == "" ||
		flags.LogLevel == "" ||
		flags.Port == "" ||
		flags.JsFiltersPath == "" ||
		flags.GoPluginsPath == ""

	requiredFlagsNames := []string{
		dockerAPIVersion,
		dockerSockURL,
		targetHostName,
		prettyLog,
		port,
		jsFiltersPath,
		goPluginsPath,
	}

	if haveEmptyRequiredFlags {
		msg := fmt.Sprintf("The following flags cannot be empty:")
		for _, name := range requiredFlagsNames {
			msg += fmt.Sprintf("\n\t%v", name)
		}
		panic(msg)
	}
}
