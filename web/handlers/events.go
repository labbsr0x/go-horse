package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/labbsr0x/go-horse/web/config-web"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/ioutils"
	"github.com/kataras/iris"
)

type EventsAPI interface {
	EventsHandler(ctx iris.Context)
}

type DefaultEventsAPI struct {
	*web.WebBuilder
}

// InitFromWebBuilder initializes a default consent api instance from a web builder instance
func (dapi *DefaultEventsAPI) InitFromWebBuilder(webBuilder *web.WebBuilder) *DefaultEventsAPI {
	dapi.WebBuilder = webBuilder
	return dapi
}

// EventsHandler handle logs command
func (dapi *DefaultEventsAPI) EventsHandler(ctx iris.Context) {

	contextWithCancel, cancel := context.WithCancel(context.Background())
	defer cancel()

	messagesChannel, errorChannel := dapi.DockerCli.Events(contextWithCancel, types.EventsOptions{})

	writer := ctx.ResponseWriter()
	ctx.ResetResponseWriter(writer)

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
		case <-contextWithCancel.Done():
			return
		case <-errorChannel:
			return
		case message := <-messagesChannel:
			enc.Encode(message)
		}
	}
}
