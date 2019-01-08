package server

import (
	"context"
	"os"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/config"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/handlers"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// GoHorse GoHorse
func GoHorse() *iris.Application {

	logSetup()

	app := iris.New()
	app.Use(recover.New())

	//TODO mapear rota para receber token ou nao
	authToken := app.Party("/token/{token:string}/")
	authToken.Post("/{version:string}/containers/{containerId:string}/attach", handlers.AtachHandler)
	authToken.Get("/{version:string}/containers/{containerId:string}/logs", handlers.LogsHandler)
	authToken.Post("/{version:string}/containers/{containerId:string}/wait", handlers.WaitHandler)
	authToken.Post("/{version:string}/exec/{execInstanceId:string}/start", handlers.ExecHandler)
	authToken.Get("/{version:string}/containers/{containerId:string}/stats", handlers.StatsHandler)

	app.Post("/{version:string}/containers/{containerId:string}/attach", handlers.AtachHandler)
	app.Get("/{version:string}/containers/{containerId:string}/logs", handlers.LogsHandler)
	app.Post("/{version:string}/containers/{containerId:string}/wait", handlers.WaitHandler)
	app.Post("/{version:string}/exec/{execInstanceId:string}/start", handlers.ExecHandler)
	app.Get("/{version:string}/containers/{containerId:string}/stats", handlers.StatsHandler)
	app.Any("*", handlers.ProxyHandler)

	app.Run(iris.Addr(config.Port), iris.WithoutStartupLog)
	return app
}

// Restart Restart
func Restart(app *iris.Application) *iris.Application {
	app.Shutdown(context.Background())
	return GoHorse()
}

func logSetup() {
	zerolog.SetGlobalLevel(config.LogLevel)
	if config.PrettyLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
