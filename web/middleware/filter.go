package middleware

import (
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/util"
	"github.com/kataras/iris/context"
	"github.com/rs/zerolog/log"
	"io/ioutil"
)


const (
	RequestBodyKey  = "requestBody"
	ResponseBodyKey = "responseBody"
)



func ResquestFilter(filter * filters.FilterManager) context.Handler {
	return func(ctx context.Context) {
		util.SetFilterContextValues(ctx)

		if ctx.Request().Body != nil {
			requestBody, err := ioutil.ReadAll(ctx.Request().Body)
			if err != nil {
				log.Error().Str("request", ctx.String()).Err(err)
			}
			ctx.Values().Set(RequestBodyKey, string(requestBody))
		}

		ctx.Values().Set("path", ctx.Request().URL.Path)

		_, err := filter.RunRequestFilters(ctx, RequestBodyKey)

		writer := ctx.ResponseWriter()
		ctx.ResetResponseWriter(writer)

		if err != nil {
			log.Error().Err(err).Msg("Error during the execution of REQUEST filters")
			writer.WriteString(err.Error())
			ctx.StopExecution()
			return
		}
		ctx.Next()
	}
}