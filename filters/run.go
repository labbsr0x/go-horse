package filters

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"

	filter "github.com/labbsr0x/go-horse/filters/config-filter"
	"github.com/labbsr0x/go-horse/filters/list"
	"github.com/labbsr0x/go-horse/filters/model"
	"github.com/labbsr0x/go-horse/prometheus"
	"github.com/kataras/iris"
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

			logrus.WithFields(logrus.Fields{
				"Filter matched": ctx.String(),
				"filter_config": fmt.Sprintf("%#v", filterConfig),
			}).Debugf("Executing %s filter %s...", filterConfig.InvokeName(), filterConfig.Name)

			start := time.Now()

			result, err = filter.Exec(ctx, ctx.Values().GetString(bodyKey))

			statusCode := strconv.Itoa(result.Status)
			metrics := prometheus.GetMetrics()
			metrics.FilterCount.WithLabelValues(filterConfig.Name, filterConfig.InvokeName(), statusCode).Inc()
			metrics.FilterLatency.WithLabelValues(filterConfig.Name, filterConfig.InvokeName(), statusCode).
				Observe(float64(time.Since(start).Seconds()) / 1000000000)

			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err.Error(),
				}).Errorf("Error applying filter : %s", filterConfig.Name)
			}

			logrus.WithFields(logrus.Fields{
				"Filter output": fmt.Sprintf("%#v", result),
				"filter_config": fmt.Sprintf("%#v", result),
			}).Debugf("Filter execution end")

			if result.Operation == model.Write {
				logrus.WithFields(logrus.Fields{
					"Filter": filterConfig.Name,
				}).Debugf("Body rewrite for filte")
				ctx.Values().Set(bodyKey, result.Body)
			}

			if !result.Next {
				logrus.WithFields(logrus.Fields{
					"Filter": filterConfig.Name,
				}).Infof("Filter chain canceled by filter")
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
