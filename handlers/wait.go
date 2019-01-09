package handlers

import (
	"context"
	"fmt"
	"io"
	"time"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/docker/docker/api/types/container"
	"github.com/kataras/iris"
)

// WaitHandler lero lero
func WaitHandler(ctx iris.Context) {

	util.SetEnvVars(ctx)

	_, er := runRequestFilters(ctx)

	if er != nil {
		ctx.StopExecution()
		return
	}

	params := ctx.FormValues()
	condition := util.GetRequestParameter(params, "condition")

	context := context.Background()
	resp, err := dockerCli.ContainerWait(context, ctx.Params().Get("containerId"), container.WaitCondition(condition))

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
		// if erroWait != nil {
		// 	log.Error().Err(erroWait).Msgf("Erro ao executar o wait")
		// }
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
