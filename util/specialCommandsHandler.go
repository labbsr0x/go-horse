package util

import (
	"regexp"

	"github.com/kataras/iris"
)

// Handler Handler
type Handler func(iris.Context)

// GoHorseSpecialHandler GoHorseSpecialHandler
type GoHorseSpecialHandler interface {
	GetURLPattern() regexp.Regexp
}

var handlers []Handler

// SetHandler SetHandler
func SetHandler(specialCommnadsHandlers ...interface{}) {
	handlers := make([]interface{}, 10)
	for _, handler := range specialCommnadsHandlers {
		_, okIris := asIrisHandler(handler)
		_, okSpecial := asSpecialHandler(handler)
		if okIris && okSpecial {
			handlers = append(handlers, handler)
		}
	}
	handlers = specialCommnadsHandlers
}

func asIrisHandler(handler interface{}) (hndlr iris.Handler, ok bool) {
	hndlr, ok = handler.(iris.Handler)
	return
}

func asSpecialHandler(handler interface{}) (hndlr GoHorseSpecialHandler, ok bool) {
	hndlr, ok = handler.(GoHorseSpecialHandler)
	return
}

// HandleNonStrictHTTPDockerCommands HandleNonStrictHTTPDockerCommands
func HandleNonStrictHTTPDockerCommands(ctx iris.Context) {

}
