package main

import (
	"os"

	"gitex.labbs.com.br/labbsr0x/proxy/sandman-acl-proxy/config"
	"gitex.labbs.com.br/labbsr0x/proxy/sandman-acl-proxy/handlers"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	app := iris.New()
	app.Use(recover.New())
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
