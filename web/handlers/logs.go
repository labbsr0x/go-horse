package handlers

import (
	"context"
	"io"
	"time"

	"github.com/labbsr0x/go-horse/web/config-web"

	"github.com/labbsr0x/go-horse/util"
	"github.com/docker/docker/api/types"
	"github.com/kataras/iris"
)

type LogsAPI interface {
	LogsHandler(ctx iris.Context)
}

type DefaultLogsAPI struct {
	*web.WebBuilder
}

// InitFromWebBuilder initializes a default consent api instance from a web builder instance
func (dapi *DefaultLogsAPI) InitFromWebBuilder(webBuilder *web.WebBuilder) *DefaultLogsAPI {
	dapi.WebBuilder = webBuilder
	return dapi
}

// LogsHandler handle logs command
func (dapi *DefaultLogsAPI) LogsHandler(ctx iris.Context) {

	params := ctx.FormValues()

	options := types.ContainerLogsOptions{
		ShowStdout: util.GetRequestParameter(params, "stdout") == "1",
		ShowStderr: util.GetRequestParameter(params, "stderr") == "1",
		Since:      util.GetRequestParameter(params, "since"),
		Until:      util.GetRequestParameter(params, "until"),
		Timestamps: util.GetRequestParameter(params, "timestamps") == "1",
		Follow:     util.GetRequestParameter(params, "follow") == "1",
		Tail:       util.GetRequestParameter(params, "tail"),
		Details:    util.GetRequestParameter(params, "details") == "1",
	}

	var responseBody io.ReadCloser
	var err error

	if ctx.GetCurrentRoute().Name() == "container-logs" {
		responseBody, err = dapi.DockerCli.ContainerLogs(context.Background(), ctx.Params().Get("id"), options)
	} else {
		responseBody, err = dapi.DockerCli.ServiceLogs(context.Background(), ctx.Params().Get("id"), options)
	}

	defer responseBody.Close()

	writer := ctx.ResponseWriter()
	if err != nil {
		writer.WriteString(err.Error())
		return
	}

	var nr int

	for {
		time.Sleep(time.Millisecond * 100)
		var buf = make([]byte, 32*1024)
		var nr2 int
		var er error
		nr2, er = responseBody.Read(buf)
		nr += nr2
		if er == io.EOF {
			writer.Write(buf)
			break
		}
		if er != nil {
			writer.Write(buf)
			break
		}
		writer.Write(buf)
	}

	writer.SetWritten(nr)

	return

}
