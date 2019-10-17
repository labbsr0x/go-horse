package handlers

import (
	"context"
	"fmt"
	"io"
	"time"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/config-web"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/kataras/iris"
)

type StatsAPI interface {
	StatsHandler(ctx iris.Context)
}

type DefaultStatsAPI struct {
	*web.WebBuilder
}

// InitFromWebBuilder initializes a default consent api instance from a web builder instance
func (dapi *DefaultStatsAPI) InitFromWebBuilder(webBuilder *web.WebBuilder) *DefaultStatsAPI {
	dapi.WebBuilder = webBuilder
	return dapi
}

// StatsHandler handle logs command
func (dapi *DefaultStatsAPI) StatsHandler(ctx iris.Context) {

	params := ctx.FormValues()

	response, err := dapi.DockerCli.ContainerStats(context.Background(), ctx.Params().Get("containerId"), util.GetRequestParameter(params, "stream") == "1")

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
