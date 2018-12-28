package main

import (
	"os"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/config"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/handlers"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	app := iris.New()
	app.Use(recover.New())

	app.Post("/{version:string}/containers/{containerId:string}/attach", handlers.AtachHandler)
	app.Get("/{version:string}/containers/{containerId:string}/logs", handlers.LogsHandler)
	app.Post("/{version:string}/containers/{containerId:string}/wait", handlers.WaitHandler)
	app.Post("/{version:string}/exec/{execInstanceId:string}/start", handlers.ExecHandler)
	app.Any("*", handlers.ProxyHandler)
	app.Run(iris.Addr(config.Port))
}

func init() {
	// zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(config.LogLevel)
	if config.PrettyLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
