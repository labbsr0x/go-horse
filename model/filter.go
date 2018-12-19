package model

import (
	"regexp"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/plugins"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/filters"
	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"
)

// BodyOperation json body operation type
type BodyOperation int

const (
	// Read read
	Read BodyOperation = 0
	// Write write
	Write BodyOperation = 1
)

// Invoke invoke
type Invoke int

const (
	// Response Response
	Response Invoke = 0
	// Request Request
	Request Invoke = 1
)

// Filter lero-lero
type Filter interface {
	Config() FilterConfig
	Exec(ctx iris.Context, requestBody string) FilterReturn
	MatchURL(ctx iris.Context) bool
}

// FilterConfig lero-lero
type FilterConfig struct {
	Name        string
	Order       int
	PathPattern string
	Invoke      Invoke
	Function    string
	regex       *regexp.Regexp
}

// FilterReturn lero-lero
type FilterReturn struct {
	Next      bool
	Body      string
	Status    int
	Operation BodyOperation
	Err       error
}

type baseFilter struct {
	Filter
	FilterConfig
}

func parseOperation(operation interface{}) BodyOperation {
	intOperation, ok := operation.(int)
	if ok {
		if intOperation == 1 {
			return Write
		}
		return Read
	}
	if operation == filters.Write {
		return Write
	}
	return Read
}

func parseInvoke(invoke interface{}) Invoke {
	intInvoke, ok := invoke.(int)
	if ok {
		if intInvoke == 1 {
			return Request
		}
		return Response
	}
	if invoke == filters.Request {
		return Request
	}
	return Response
}

// FilterGO lero-lero
type FilterGO struct {
	baseFilter
	innerType plugins.Filter
}

// FilterJS lero-lero
type FilterJS struct {
	baseFilter
	innerType filters.JsFilterModel
}

// NewFilterJS lero-lero
func NewFilterJS(innerType filters.JsFilterModel) FilterJS {
	filterJs := FilterJS{}
	filterJs.innerType = innerType
	filterJs.FilterConfig = FilterConfig{Name: innerType.Name, Order: innerType.Order, PathPattern: innerType.PathPattern, Invoke: parseInvoke(innerType.Invoke), Function: innerType.Function}
	return filterJs
}

// MatchURL lero-lero
func (filterJs FilterJS) MatchURL(ctx iris.Context) bool {
	return MatchURL(ctx, filterJs.baseFilter)
}

// Config lero-lero
func (filterJs FilterJS) Config() FilterConfig {
	return filterJs.FilterConfig
}

// Exec lero-lero
func (filterJs FilterJS) Exec(ctx iris.Context, requestBody string) FilterReturn {
	jsReturn := filterJs.innerType.Exec(ctx, requestBody)
	return FilterReturn{jsReturn.Next, jsReturn.Body, jsReturn.Status, parseOperation(jsReturn.Operation), jsReturn.Err}
}

// MatchURL lero-lero
func (filterGo FilterGO) MatchURL(ctx iris.Context) bool {
	return MatchURL(ctx, filterGo.baseFilter)
}

// Config lero-lero
func (filterGo FilterGO) Config() FilterConfig {
	return filterGo.FilterConfig
}

// Exec lero-lero
func (filterGo FilterGO) Exec(ctx iris.Context, requestBody string) FilterReturn {
	Next, Body, Status, Operation, Err := filterGo.innerType.Exec(ctx, requestBody)
	return FilterReturn{Next, Body, Status, parseOperation(Operation), Err}
}

// NewFilterGO lero-lero
func NewFilterGO(innerType plugins.Filter) FilterGO {
	filterGo := FilterGO{}
	filterGo.innerType = innerType
	Name, Order, PathPattern, Invoke := innerType.Config()
	filterGo.FilterConfig = FilterConfig{Name: Name, Order: Order, PathPattern: PathPattern, Invoke: parseInvoke(Invoke), Function: ""}
	return filterGo
}

// MatchURL lero lero
func MatchURL(ctx iris.Context, base baseFilter) bool {
	if base.regex == nil {
		regex, error := regexp.Compile(base.PathPattern)
		if error != nil {
			log.Error().Str("plugin_name", base.Name).Err(error).Msg("Error compiling the filter url matcher regex")
		} else {
			base.regex = regex
		}
	}
	return base.regex.MatchString(ctx.RequestPath(false))
}
