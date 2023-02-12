package metrics

import (
	"context"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	clientmetrics "k8s.io/client-go/tools/metrics"
)

const (
	Subsystem  = "KwB"
	RestClientSubsystem = "rest_client"
	LatencyKey          = "request_latency_seconds"
	ResultKey           = "requests_total"
)

var (
	RequestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem: RestClientSubsystem,
		Name:      LatencyKey,
		Help:      "Request latency in seconds. Broken down by verb and URL.",
		Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
	}, []string{"verb", "url"})

	requestResult = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: RestClientSubsystem,
		Name:      ResultKey,
		Help:      "Number of HTTP requests, partitioned by status code, method, and host.",
	}, []string{"code", "method", "host"})

	errorCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   Subsystem,
			Name:        "operation_error_total",
			Help:        "Total errors per resource.",
			ConstLabels: nil,
		},
		[]string{"resource", "name", "verb"},
	)

	requests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   Subsystem,
			Name:        "operation_request_total",
			Help:        "Total request per resource.",
			ConstLabels: nil,
		},
		[]string{"resource", "name", "verb"},
	)
)

func ResourceErrors(resource string, name string, verb string) {
	errorCount.WithLabelValues(resource, name, verb).Inc()
}

func ResourceCount(resource string, name string, verb string) {
	requests.WithLabelValues(resource, name, verb).Inc()
}

// The below is a re-implemntation of https://github.com/kubernetes-sigs/controller-runtime/blob/17893a8fae1e32209b8c8caa52f2ae6b32d6f6de/pkg/metrics/client_go_adapter.go#L42
type resultAdapter struct {
	metric *prometheus.CounterVec
}

func (r *resultAdapter) Increment(_ context.Context, code, method, host string) {
	r.metric.WithLabelValues(code, method, host).Inc()
}

// LatencyAdapter implements LatencyMetric.
type LatencyAdapter struct {
	metric *prometheus.HistogramVec
}

// Observe increments the request latency metric for the given verb/URL.
func (l *LatencyAdapter) Observe(_ context.Context, verb string, u url.URL, latency time.Duration) {
	l.metric.WithLabelValues(verb, u.String()).Observe(latency.Seconds())
}

func init() {
	prometheus.MustRegister(RequestLatency)
	prometheus.MustRegister(requestResult)
	clientmetrics.Register(clientmetrics.RegisterOpts{
		ClientCertExpiry:      nil,
		ClientCertRotationAge: nil,
		RequestLatency:        &LatencyAdapter{RequestLatency},
		RequestSize:           nil,
		ResponseSize:          nil,
		RateLimiterLatency:    nil,
		RequestResult:         &resultAdapter{requestResult},
		ExecPluginCalls:       nil,
	})
}
