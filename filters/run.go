package filters

import (
	"fmt"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/list"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/model"
	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"
	"net/http"
)

func RunRequestFilters(ctx iris.Context, requestBodyKey string) (result model.FilterReturn, err error) {
	return runFilters(ctx, requestBodyKey, list.RequestFilters())
}

func RunResponseFilters(ctx iris.Context, responseBodyKey string) (result model.FilterReturn, err error) {
	return runFilters(ctx, responseBodyKey, list.ResponseFilters())
}

func runFilters(ctx iris.Context, bodyKey string, filters []model.Filter) (result model.FilterReturn, err error) {
	for _, filter := range filters {
		if filter.MatchURL(ctx) {
			filterConfig := filter.Config()
			log.Debug().Str("Filter matched", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", filterConfig)).Msgf("Executing %s filter %s...", filterConfig.InvokeName(), filterConfig.Name)
			result, err = filter.Exec(ctx, ctx.Values().GetString(bodyKey))
			if err != nil {
				log.Error().Err(err).Msgf("Error applying filter : %s", filterConfig.Name)
			}
			log.Debug().Str("Filter output", fmt.Sprintf("%#v", result)).Str("filter_config", fmt.Sprintf("%#v", result)).Msg("filter execution end")
			if result.Operation == model.Write {
				log.Debug().Msgf("Body rewrite for filter : %s", filterConfig.Name)
				ctx.Values().Set(bodyKey, result.Body)
			}
			if !result.Next {
				log.Info().Msgf("Filter chain canceled by filter - %s", filterConfig.Name)
				break
			}
		}
	}

	if err != nil {
		if result.Status == 0 {
			result.Status = http.StatusInternalServerError
		}
		ctx.StatusCode(result.Status)
		ctx.ContentType("application/json")
		ctx.WriteString(err.Error())
	}

	return
}
