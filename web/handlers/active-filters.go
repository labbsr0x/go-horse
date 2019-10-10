package handlers

import (
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/list"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/config"
	"github.com/kataras/iris"
)

type ActiveFiltersAPI interface {
	ActiveFiltersHandler(ctx iris.Context)
}

type DefaultActiveFiltersAPI struct {
	*config.WebBuilder
}

// InitFromWebBuilder initializes a default consent api instance from a web builder instance
func (dapi *DefaultActiveFiltersAPI) InitFromWebBuilder(webBuilder *config.WebBuilder) *DefaultActiveFiltersAPI {
	dapi.WebBuilder = webBuilder
	return dapi
}

// ActiveFiltersHandler ActiveFiltersHandler
func (dapi *DefaultActiveFiltersAPI) ActiveFiltersHandler(ctx iris.Context) {
	_, _ = ctx.JSON(iris.Map{
		"request":  list.RequestFilters(),
		"response": list.ResponseFilters(),
	})
}
