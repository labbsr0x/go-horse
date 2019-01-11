package handlers

import (
	"context"
	"fmt"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/docker/docker/api/types"
	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"
)

// AttachHandler handle attach command
func AttachHandler(ctx iris.Context) {

	util.SetFilterContextValues(ctx)

	_, err := filters.RunRequestFilters(ctx, RequestBodyKey)

	if err != nil {
		ctx.StopExecution()
		return
	}

	params := ctx.FormValues()

	context := context.Background()
	options := types.ContainerAttachOptions{}

	options.Stream = util.GetRequestParameter(params, "stream") == "1"
	options.Stdin = util.GetRequestParameter(params, "stdin") == "1"
	options.Stdout = util.GetRequestParameter(params, "stdout") == "1"
	options.Stderr = util.GetRequestParameter(params, "stderr") == "1"
	options.DetachKeys = util.GetRequestParameter(params, "detachKeys")
	options.Logs = util.GetRequestParameter(params, "logs") == "1"

	resp, err := dockerCli.ContainerAttach(context, ctx.Params().Get("containerId"), options)

	if err != nil {
		log.Error().Err(err).Msg("Error executing docker client # ContainerExecAttach")
	}

	msgs := make(chan []byte)
	msgsErr := make(chan error)

	go func() {
		for {
			msg, er := resp.Reader.ReadBytes('\n')
			if er != nil {
				msgsErr <- er
				return
			}
			msgs <- msg
		}
	}()

	_, upgrade := ctx.Request().Header["Upgrade"]
	writer := ctx.ResponseWriter()
	ctx.ResetResponseWriter(writer)
	conn, _, err := writer.Hijack()
	if err != nil {
		log.Error().Err(err).Msg("conn hijack failed")
	}

	conn.Write([]byte{})

	if upgrade {
		fmt.Fprintf(conn, "HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n")
	} else {
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.docker.raw-stream\r\n\r\n")
	}

msgLoop:
	for {
		select {
		case msg := <-msgs:
			fmt.Fprintf(conn, "%s", msg)
		case <-msgsErr:
			break msgLoop
		}
	}

	defer close(msgs)
	defer close(msgsErr)
	defer conn.Close()
	defer resp.Close()

	ctx.StopExecution()
	ctx.EndRequest()
	return
}
