package handlers

import (
	"context"
	"fmt"
	"io"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/docker/docker/api/types"
	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"
)

// ExecHandler handle the exec command
func ExecHandler(ctx iris.Context) {

	util.SetEnvVars(ctx)

	_, err := runRequestFilters(ctx)

	if err != nil {
		ctx.StopExecution()
		return
	}

	var execStartCheck types.ExecStartCheck

	if err := ctx.ReadJSON(&execStartCheck); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.WriteString(err.Error())
		return
	}

	context := context.Background()

	resp, err := dockerCli.ContainerExecAttach(context, ctx.Params().Get("execInstanceId"), execStartCheck)
	if err != nil {
		log.Error().Err(err).Msg("Error executing docker client # ContainerExecAttach")
	}
	defer resp.Close()

	msgs := make(chan []byte)
	msgsErr := make(chan error)
	defer close(msgs)
	defer close(msgsErr)

	go func() {
		for {
			msg, er := resp.Reader.ReadByte()
			if er != nil {
				msgsErr <- er
				return
			}
			msgs <- []byte{msg}
		}
	}()

	_, upgrade := ctx.Request().Header["Upgrade"]

	conn, _, err := ctx.ResponseWriter().Hijack()
	if err != nil {
		log.Error().Err(err).Msg("conn hijack failed")
	}

	conn.Write([]byte{})
	resp.Conn.Write([]byte{})

	if upgrade {
		fmt.Fprintf(conn, "HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n")
	} else {
		fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.docker.raw-stream\r\n\r\n")
	}

	go func() {
		var nr int
		for {
			var buf = make([]byte, 1)
			var nr2 int
			var er error
			nr2, er = conn.Read(buf)
			nr += nr2
			if er == io.EOF {
				resp.Conn.Write(buf)
				break
			}
			if er != nil {
				break
			}
			resp.Conn.Write(buf)
		}
	}()

msgLoop:
	for {
		select {
		case msg := <-msgs:
			fmt.Fprintf(conn, "%s", msg)
		case <-msgsErr:
			defer conn.Close()
			break msgLoop
		}
	}
	ctx.StopExecution()
	ctx.EndRequest()
	return

}
