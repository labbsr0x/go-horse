package web

import (
	stdContext "context"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/prometheus"
	web "gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/config-web"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/handlers"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/middleware"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server holds the information needed to run Whisper
type Server struct {
	*web.WebBuilder
	ActiveFiltersAPIs handlers.ActiveFiltersAPI
	AttachAPIs        handlers.AttachAPI
	LogsAPIs          handlers.LogsAPI
	WaitAPIs          handlers.WaitAPI
	ExecAPIs          handlers.ExecAPI
	StatsAPIs         handlers.StatsAPI
	EventsAPIs        handlers.EventsAPI
	ProxyAPIs         handlers.ProxyAPI
}

// InitFromWebBuilder builds a Server instance
func (s *Server) InitFromWebBuilder(webBuilder *web.WebBuilder) *Server {
	s.WebBuilder = webBuilder
	s.ActiveFiltersAPIs = new(handlers.DefaultActiveFiltersAPI).InitFromWebBuilder(webBuilder)
	s.AttachAPIs = new(handlers.DefaultAttachAPI).InitFromWebBuilder(webBuilder)
	s.LogsAPIs = new(handlers.DefaultLogsAPI).InitFromWebBuilder(webBuilder)
	s.WaitAPIs = new(handlers.DefaultWaitAPI).InitFromWebBuilder(webBuilder)
	s.ExecAPIs = new(handlers.DefaultExecAPI).InitFromWebBuilder(webBuilder)
	s.StatsAPIs = new(handlers.DefaultStatsAPI).InitFromWebBuilder(webBuilder)
	s.EventsAPIs = new(handlers.DefaultEventsAPI).InitFromWebBuilder(webBuilder)
	s.ProxyAPIs = new(handlers.DefaultProxyAPI).InitFromWebBuilder(webBuilder)

	logLevel, err := logrus.ParseLevel(s.LogLevel)
	if err != nil {
		logrus.Errorf("Not able to parse log level string. Setting default level: info.")
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)

	return s
}

// Run initializes the web server and its apis
func (s *Server) Run() error {

	app := iris.New()
	app.Use(recover.New())
	app.Use(prometheus.GetMetrics().ServeHTTP)

	app.Get("/active-filters", s.ActiveFiltersAPIs.ActiveFiltersHandler)
	app.Get("/metrics", iris.FromStd(promhttp.Handler()))

	app.Use(middleware.ResquestFilter(s.Filter))

	//TODO mapear rota para receber token ou nao
	authToken := app.Party("/token/{token:string}/")
	authToken.Post("/{version:string}/containers/{containerId:string}/attach", s.AttachAPIs.AttachHandler)
	authToken.Get("/{version:string}/containers/{id:string}/logs", s.LogsAPIs.LogsHandler).Name = "container-logs"
	authToken.Get("/{version:string}/services/{id:string}/logs", s.LogsAPIs.LogsHandler).Name = "service-logs"
	authToken.Post("/{version:string}/containers/{containerId:string}/wait", s.WaitAPIs.WaitHandler)
	authToken.Post("/{version:string}/exec/{execInstanceId:string}/start", s.ExecAPIs.ExecHandler)
	authToken.Get("/{version:string}/containers/{containerId:string}/stats", s.StatsAPIs.StatsHandler)
	authToken.Get("/{version:string}/events", s.EventsAPIs.EventsHandler)

	app.Post("/{version:string}/containers/{containerId:string}/attach", s.AttachAPIs.AttachHandler)
	app.Get("/{version:string}/containers/{id:string}/logs", s.LogsAPIs.LogsHandler).Name = "container-logs"
	app.Get("/{version:string}/services/{id:string}/logs", s.LogsAPIs.LogsHandler).Name = "service-logs"
	app.Post("/{version:string}/containers/{containerId:string}/wait", s.WaitAPIs.WaitHandler)
	app.Post("/{version:string}/exec/{execInstanceId:string}/start", s.ExecAPIs.ExecHandler)
	app.Get("/{version:string}/containers/{containerId:string}/stats", s.StatsAPIs.StatsHandler)
	app.Get("/{version:string}/events", s.EventsAPIs.EventsHandler)
	app.Any("*", s.ProxyAPIs.ProxyHandler)

	return s.ListenAndServe(app)
}

func (s *Server) ListenAndServe(app *iris.Application) error {

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch,
			os.Interrupt,
			syscall.SIGINT,
			os.Kill,
			syscall.SIGKILL,
			syscall.SIGTERM,
		)
		<-ch
		logrus.Info("Server Stopped")
		logrus.Debugf("Waiting %v seconds",time.Second * s.ShutdownTime)
		ctx, cancel := stdContext.WithTimeout(stdContext.Background(), time.Second * s.ShutdownTime)

		defer cancel()

		if err := app.Shutdown(ctx); err != nil{
			logrus.Fatal("server finalization error: %v", err)
		}

		logrus.Info("Server Exited Properly")
	}()

	logrus.Infof("Starting Server")
	return app.Run(iris.Addr(s.Flags.Port), iris.WithoutInterruptHandler)
}