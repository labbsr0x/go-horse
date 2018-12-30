package model

import (
	"regexp"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/plugins"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters"
	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"
)

// BodyOperation : Read or Write > If the body content was updated by the filter and needs to be overwritten in the response,  Write property should be passed
type BodyOperation int

const (
	// Read : no changes, keep the original
	Read BodyOperation = 0
	// Write : changes made, override the old
	Write BodyOperation = 1
)

// Invoke : a property to tell if the filter is gonna be executed before (Request) or after (Response) the client request be send to the docker daemon
type Invoke int

const (
	// Response filter invoke on the response from the docker daemon
	Response Invoke = 0
	// Request filter invoke on the request from the docker client
	Request Invoke = 1
)

// Filter common filter interface between go and javascript filters
type Filter interface {
	Config() FilterConfig
	Exec(ctx iris.Context, requestBody string) (FilterReturn, error)
	MatchURL(ctx iris.Context) bool
}

// FilterConfig common filter configuration
type FilterConfig struct {
	Name        string
	Order       int
	PathPattern string
	Invoke      Invoke
	Function    string
	regex       *regexp.Regexp
}

// FilterReturn common filter return
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
		return BodyOperation(intOperation)
	}
	if operation == filters.Write {
		return Write
	}
	return Read
}

func parseInvoke(invoke interface{}) Invoke {
	intInvoke, ok := invoke.(int)
	if ok {
		return Invoke(intInvoke)
	}
	if invoke == filters.Request {
		return Request
	}
	return Response
}

// FilterGO Go proxy filter
type FilterGO struct {
	baseFilter
	innerType plugins.Filter
}

// FilterJS JS proxy filter
type FilterJS struct {
	baseFilter
	innerType filters.JsFilterModel
}

// NewFilterJS JS filter factory
func NewFilterJS(innerType filters.JsFilterModel) FilterJS {
	filterJs := FilterJS{}
	filterJs.innerType = innerType
	filterJs.FilterConfig = FilterConfig{Name: innerType.Name, Order: innerType.Order, PathPattern: innerType.PathPattern, Invoke: parseInvoke(innerType.Invoke), Function: innerType.Function}
	return filterJs
}

// MatchURL js
func (filterJs FilterJS) MatchURL(ctx iris.Context) bool {
	return MatchURL(ctx, filterJs.baseFilter)
}

// Config js
func (filterJs FilterJS) Config() FilterConfig {
	return filterJs.FilterConfig
}

// Exec js
func (filterJs FilterJS) Exec(ctx iris.Context, requestBody string) (FilterReturn, error) {
	jsReturn, error := filterJs.innerType.Exec(ctx, requestBody)
	if error != nil {
		return FilterReturn{}, error
	}
	return FilterReturn{jsReturn.Next, jsReturn.Body, jsReturn.Status, parseOperation(jsReturn.Operation), jsReturn.Err}, nil
}

// MatchURL go
func (filterGo FilterGO) MatchURL(ctx iris.Context) bool {
	return MatchURL(ctx, filterGo.baseFilter)
}

// Config go
func (filterGo FilterGO) Config() FilterConfig {
	return filterGo.FilterConfig
}

// Exec go
func (filterGo FilterGO) Exec(ctx iris.Context, requestBody string) (FilterReturn, error) {
	Next, Body, Status, Operation, Err := filterGo.innerType.Exec(ctx, requestBody)
	return FilterReturn{Next, Body, Status, parseOperation(Operation), Err}, Err
}

// NewFilterGO filter factory
func NewFilterGO(innerType plugins.Filter) FilterGO {
	filterGo := FilterGO{}
	filterGo.innerType = innerType
	Name, Order, PathPattern, Invoke := innerType.Config()
	filterGo.FilterConfig = FilterConfig{Name: Name, Order: Order, PathPattern: PathPattern, Invoke: parseInvoke(Invoke), Function: ""}
	return filterGo
}

// MatchURL common
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
