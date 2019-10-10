package handlers

import (
	"context"
	"encoding/json"
	"time"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/config"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/kataras/iris"
)

type EventsAPI interface {
	EventsHandler(ctx iris.Context)
}

type DefaultEventsAPI struct {
	*config.WebBuilder
}

// InitFromWebBuilder initializes a default consent api instance from a web builder instance
func (dapi *DefaultEventsAPI) InitFromWebBuilder(webBuilder *config.WebBuilder) *DefaultEventsAPI {
	dapi.WebBuilder = webBuilder
	return dapi
}

// EventsHandler handle logs command
func (dapi *DefaultEventsAPI) EventsHandler(ctx iris.Context) {

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
