package handlers

import (
	"context"
	"encoding/json"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters"
	"time"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/kataras/iris"
)

// EventsHandler handle logs command
func EventsHandler(ctx iris.Context) {

	util.SetFilterContextValues(ctx)

	_, err := filters.RunRequestFilters(ctx, RequestBodyKey)

	if err != nil {
		ctx.StopExecution()
		return
	}

	context, cancel := context.WithCancel(context.Background())
	defer cancel()

	messagesChannel, errorChannel := dockerCli.Events(context, types.EventsOptions{})

	writer := ctx.ResponseWriter()
	ctx.ResetResponseWriter(writer)

	if err != nil {
		writer.WriteString(err.Error())
		return
	}

	writer.Header().Set("Content-Type", "application/json")

	output := ioutils.NewWriteFlusher(writer)
	defer output.Close()
	output.Flush()

	enc := json.NewEncoder(output)

	// TODO FIX ME
	timeout := time.After(1 * time.Minute)

	for {
		select {
		case <-timeout:
			return
		case <-context.Done():
			return
		case <-errorChannel:
			return
		case message := <-messagesChannel:
			enc.Encode(message)
		}
	}
}
