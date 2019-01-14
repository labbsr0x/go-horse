package handlers

import (
	"context"
	"fmt"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters"
	"io"
	"time"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/kataras/iris"
)

// StatsHandler handle logs command
func StatsHandler(ctx iris.Context) {

	util.SetFilterContextValues(ctx)

	_, err := filters.RunRequestFilters(ctx, RequestBodyKey)

	if err != nil {
		ctx.StopExecution()
		return
	}

	context := context.Background()

	params := ctx.FormValues()

	response, err := dockerCli.ContainerStats(context, ctx.Params().Get("containerId"), util.GetRequestParameter(params, "stream") == "1")

	writer := ctx.ResponseWriter()
	ctx.ResetResponseWriter(writer)

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
		nr2, er = response.Body.Read(buf)
		nr += nr2
		if er == io.EOF {
			writer.Write(buf[:nr2])
			break
		}
		if er != nil {
			fmt.Println(0, er)
		}
		writer.Write(buf[:nr2])
	}

	writer.SetWritten(nr)

	return

}
