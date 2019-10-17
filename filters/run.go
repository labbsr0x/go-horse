package filters

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	filter "gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/config-filter"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/list"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/model"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/prometheus"
	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"
)

type FilterManager struct {
	*filter.FilterBuilder
	ListAPIs list.ListAPI
}

// InitFromFilterBuilder builds a Filter instance
func (f  *FilterManager) InitFromFilterBuilder(filterBuilder *filter.FilterBuilder)  *FilterManager {
	f.FilterBuilder = filterBuilder
	f.ListAPIs = new(list.DefaultListAPI).InitFromFilterBuilder(filterBuilder)
	return f
}

func (f  *FilterManager) RunRequestFilters(ctx iris.Context, requestBodyKey string) (result model.FilterReturn, err error) {
	return f.runFilters(ctx, requestBodyKey, f.ListAPIs.RequestFilters())
}

func (f  *FilterManager) RunResponseFilters(ctx iris.Context, responseBodyKey string) (result model.FilterReturn, err error) {
	return f.runFilters(ctx, responseBodyKey, f.ListAPIs.ResponseFilters())
}

func (f  *FilterManager) runFilters(ctx iris.Context, bodyKey string, filters []model.Filter) (result model.FilterReturn, err error) {
	for _, filter := range filters {
		if filter.MatchURL(ctx) {
			filterConfig := filter.Config()
			log.Debug().Str("Filter matched", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", filterConfig)).Msgf("Executing %s filter %s...", filterConfig.InvokeName(), filterConfig.Name)

			start := time.Now()

			result, err = filter.Exec(ctx, ctx.Values().GetString(bodyKey))

			statusCode := strconv.Itoa(result.Status)
			metrics := prometheus.GetMetrics()
			metrics.FilterCount.WithLabelValues(filterConfig.Name, filterConfig.InvokeName(), statusCode).Inc()
			metrics.FilterLatency.WithLabelValues(filterConfig.Name, filterConfig.InvokeName(), statusCode).
				Observe(float64(time.Since(start).Seconds()) / 1000000000)

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
