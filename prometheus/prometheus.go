package prometheus

import (
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/config"
	"strconv"
	"time"

	"github.com/kataras/iris/context"
	"github.com/prometheus/client_golang/prometheus"
)

// Prometheus is a handler that exposes prometheus metrics for the number of requests,
// the reqLatency and the response size, partitioned by status code, method and HTTP path.
//
// Usage: pass its `ServeHTTP` to a route or globally.
type MetricsPrometheus struct {
	reqCount      *prometheus.CounterVec
	reqLatency    *prometheus.HistogramVec
	FilterCount   *prometheus.CounterVec
	FilterLatency *prometheus.HistogramVec
}

var name = "go-horse"
var metrics *MetricsPrometheus

func GetMetrics() *MetricsPrometheus {
	if metrics == nil {
		metrics = &MetricsPrometheus{}
		registerMetrics(metrics)
	}
	return metrics
}

// New returns a new prometheus middleware.
//
// If buckets are empty then `DefaultBuckets` are setted.
func registerMetrics(p *MetricsPrometheus) {
	constLabels := prometheus.Labels{"service": name, "service_version": config.Version}
	p.reqCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: constLabels,
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(p.reqCount)

	p.reqLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "http_request_duration_seconds",
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
		ConstLabels: constLabels,
	},
		[]string{"code", "method", "path"},
	)

	prometheus.MustRegister(p.reqLatency)

	p.FilterCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "filter_process_total",
			Help:        "How many filters processed, partitioned by name, invoke time and status code.",
			ConstLabels: constLabels,
		},
		[]string{"name", "invoke_time", "code"},
	)
	prometheus.MustRegister(p.FilterCount)

	p.FilterLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        "filter_process_duration_seconds",
		Help:        "How long it took to process the filter, partitioned by name, invoke time and status code.",
		ConstLabels: constLabels,
	},
		[]string{"name", "invoke_time", "code"},
	)

	prometheus.MustRegister(p.FilterLatency)
}

func (p *MetricsPrometheus) ServeHTTP(ctx context.Context) {
	if ctx.Request().URL.Path == "/metrics" || ctx.Request().URL.Path == "/favicon.ico" {
		ctx.Next()
		return
	}
	start := time.Now()
	ctx.Next()
	r := ctx.Request()
	statusCode := strconv.Itoa(ctx.GetStatusCode())
	path := ctx.GetCurrentRoute().Path()

	p.reqCount.WithLabelValues(statusCode, r.Method, path).
		Inc()

	p.reqLatency.WithLabelValues(statusCode, r.Method, path).
		Observe(float64(time.Since(start).Seconds()) / 1000000000)
}
