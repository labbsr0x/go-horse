package filter

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// All envs that GHP need to work with
const (
	jsFiltersPath = "js-filters-path"
	goPluginsPath = "go-plugins-path"
)

// Flags define the fields that will be passed via cmd
type FlagsFilter struct {
	JsFiltersPath string
	GoPluginsPath string
}

// FilterBuilder defines the parametric information of a go horse filters instance
type FilterBuilder struct {
	*FlagsFilter
}

// AddFlags adds flags for Builder.
// TODO : Discuss shortcut name
func AddFlags(flags *pflag.FlagSet) {
	flags.StringP(jsFiltersPath, "j", "", "Sets the path to json filters")
	flags.StringP(goPluginsPath, "g", "", "Sets the path to go plugins")
}

// InitFromFilterBuilder initializes the web server builder with properties retrieved from Viper.
func (b *FilterBuilder) InitFromViper(v *viper.Viper) *FilterBuilder {

	flags := new(FlagsFilter)
	flags.JsFiltersPath = v.GetString(jsFiltersPath)
	flags.GoPluginsPath = v.GetString(goPluginsPath)

	flags.check()

	b.FlagsFilter = flags

	return b
}

func (flags *FlagsFilter) check() {

	logrus.Infof("Flags: '%v'", flags)

	haveEmptyRequiredFlags := flags.JsFiltersPath == "" || flags.GoPluginsPath == ""

	requiredFlagsNames := []string{
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
