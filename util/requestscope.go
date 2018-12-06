package util

import (
	"github.com/kataras/iris"
)

// Values returns the current "user" storage.
// Named path parameters and any optional data can be saved here.
// This storage, as the whole Context, is per-request lifetime.
// You can use this function to Set and Get local values
// that can be used to share information between handlers and middleware.

// RequestScopeGet lero-lero
func RequestScopeGet(ctx iris.Context, key string) string {
	return ctx.Values().GetString(key)
}

// RequestScopeSet lero-lero
func RequestScopeSet(ctx iris.Context, key, value string) {
	ctx.Values().Set(key, value)
}

// RequestScopeList lero-lero
func RequestScopeList(ctx iris.Context) map[string]string {
	values := make(map[string]string)
	ctx.Values().Visit(
		func(key string, value interface{}) {
			values[key] = ctx.Values().GetString(key)
		})
	return values
}
