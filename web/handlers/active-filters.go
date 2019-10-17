package handlers

import (
	web "gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/config-web"
	"github.com/kataras/iris"
)

type ActiveFiltersAPI interface {
	ActiveFiltersHandler(ctx iris.Context)
}

type DefaultActiveFiltersAPI struct {
	*web.WebBuilder
}

// InitFromWebBuilder initializes a default consent api instance from a web builder instance
func (dapi *DefaultActiveFiltersAPI) InitFromWebBuilder(webBuilder *web.WebBuilder) *DefaultActiveFiltersAPI {
	dapi.WebBuilder = webBuilder
	return dapi
}

// ActiveFiltersHandler ActiveFiltersHandler
func (dapi *DefaultActiveFiltersAPI) ActiveFiltersHandler(ctx iris.Context) {
	_, _ = ctx.JSON(iris.Map{
		"request":  dapi.Filter.ListAPIs.RequestFilters(),
		"response": dapi.Filter.ListAPIs.ResponseFilters(),
	})
}
