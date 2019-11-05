package util

import (
	"os"
	"runtime"
	"strings"

	"github.com/labbsr0x/go-horse/version"

	"github.com/kataras/iris"
)

// Values returns the current "user" storage.
// Named path parameters and any optional data can be saved here.
// This storage, as the whole Context, is per-request lifetime.
// You can use this function to Set and Get local values
// that can be used to share information between handlers and middleware >> AND THE PROXY FILTERS

// RequestScopeGet RequestScopeGet
func RequestScopeGet(ctx iris.Context, key string) string {
	return ctx.Values().GetString(key)
}

// RequestScopeSet RequestScopeSet
func RequestScopeSet(ctx iris.Context, key, value string) {
	ctx.Values().Set(key, value)
}

// RequestScopeList RequestScopeList
func RequestScopeList(ctx iris.Context) map[string]string {
	values := make(map[string]string)
	ctx.Values().Visit(
		func(key string, value interface{}) {
			values[key] = ctx.Values().GetString(key)
		})
	return values
}

func SetFilterContextValues(ctx iris.Context) {
	setEnvVars(ctx)
	setConfigVars(ctx)
}

// SetEnvVars SetEnvVars
func setEnvVars(ctx iris.Context) {
	for _, env := range os.Environ() {
		pair := strings.Split(env, "=")
		ctx.Values().Set("ENV_"+pair[0], pair[1])
	}
}

func setConfigVars(ctx iris.Context) {

	ctx.Values().Set("CONFIG_VERSION", version.Version)
	ctx.Values().Set("CONFIG_GIT_COMMIT", version.GitCommit)
	ctx.Values().Set("CONFIG_BUILD_TIME", version.BuildTime)
	ctx.Values().Set("CONFIG_GO_VERSION", runtime.Version())
	ctx.Values().Set("CONFIG_OS", runtime.GOOS)
	ctx.Values().Set("CONFIG_ARCH", runtime.GOARCH)
}
