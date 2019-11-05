package handlers

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/labbsr0x/go-horse/web/config-web"

	"github.com/labbsr0x/go-horse/util"
	"github.com/docker/docker/api/types/container"
	"github.com/kataras/iris"
)

type WaitAPI interface {
	WaitHandler(ctx iris.Context)
}

type DefaultWaitAPI struct {
	*web.WebBuilder
}

// InitFromWebBuilder initializes a default consent api instance from a web builder instance
func (dapi *DefaultWaitAPI) InitFromWebBuilder(webBuilder *web.WebBuilder) *DefaultWaitAPI {
	dapi.WebBuilder = webBuilder
	return dapi
}

// WaitHandler lero lero
func (dapi *DefaultWaitAPI) WaitHandler(ctx iris.Context) {

	params := ctx.FormValues()
	condition := util.GetRequestParameter(params, "condition")

	resp, err := dapi.DockerCli.ContainerWait(context.Background(), ctx.Params().Get("containerId"), container.WaitCondition(condition))

	var respostaWait container.ContainerWaitOKBody
	var erroWait error
	finish := false

	go func() {
		select {
		case result := <-resp:
			respostaWait = result
			finish = true
		case err0 := <-err:
			erroWait = err0
			finish = true
		}
	}()

	ctx.ResetResponseWriter(ctx.ResponseWriter())

	ctx.ContentType("application/json")
	ctx.Header("Transfer-Encoding", "chunked")

	ctx.StreamWriter(func(w io.Writer) bool {
		time.Sleep(time.Second / 2)
		if finish || erroWait != nil {
			if respostaWait.Error != nil {
				fmt.Fprintf(w, "{\"StatusCode\": %d, \"Error\": {\"Message\": \"%s\"}}", respostaWait.StatusCode, respostaWait.Error.Message)
			}
			fmt.Fprintf(w, "{\"StatusCode\": %d}", respostaWait.StatusCode)
			defer func() { ctx.Next() }()
			return false
		}
		return true
	})
	ctx.StopExecution()
	return

}
