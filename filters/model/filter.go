package model

import (
	"regexp"

	"github.com/kataras/iris"
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
	Regex       *regexp.Regexp
}

// FilterReturn common filter return
type FilterReturn struct {
	Next      bool
	Body      string
	Status    int
	Operation BodyOperation
	Err       error
}

func (fc *FilterConfig) InvokeName() string {
	if fc.Invoke == Request {
		return "REQUEST"
	} else {
		return "RESPONSE"
	}
}
