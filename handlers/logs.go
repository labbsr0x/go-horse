package handlers

import (
	"context"
	"io"
	"time"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/docker/docker/api/types"
	"github.com/kataras/iris"
)

// LogsHandler handle logs command
func LogsHandler(ctx iris.Context) {

	util.SetEnvVars(ctx)

	_, err := runRequestFilters(ctx)

	if err != nil {
		ctx.StopExecution()
		return
	}

	context := context.Background()

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

	if ctx.GetCurrentRoute().Name() == "container-logs" {
		responseBody, err = dockerCli.ContainerLogs(context, ctx.Params().Get("id"), options)
	} else {
		responseBody, err = dockerCli.ServiceLogs(context, ctx.Params().Get("id"), options)
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
