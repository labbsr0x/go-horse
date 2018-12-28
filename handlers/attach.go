package handlers

import (
	"context"
	"fmt"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/docker/docker/api/types"
	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"
)

// AtachHandler lero lero
func AtachHandler(ctx iris.Context) {

	util.SetEnvVars(ctx)

	_, err := runRequestFilters(ctx)

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

	conn, _, err := ctx.ResponseWriter().Hijack()
	if err != nil {
		fmt.Println("ERRO >>>>>>>>>>>>>> ", err)
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
			defer conn.Close()
			defer resp.Close()
			break msgLoop
		}
	}
	ctx.StopExecution()
	ctx.EndRequest()
	return
}
