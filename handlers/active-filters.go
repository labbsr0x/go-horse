package handlers

import (
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/list"
	"github.com/kataras/iris"
)

// ActiveFiltersHandler ActiveFiltersHandler
func ActiveFiltersHandler(ctx iris.Context) {
	_, _ = ctx.JSON(iris.Map{
		"request":  list.RequestFilters(),
		"response": list.ResponseFilters(),
	})
}
