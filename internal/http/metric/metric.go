package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	method   = "method"
	endpoint = "endpoint"
)

type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	InflightRequests prometheus.Gauge
	RequestDuration  *prometheus.HistogramVec
}

func New() *Metrics {
	return &Metrics{
		RequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		}, []string{method, endpoint}),
		InflightRequests: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "http_inflight_requests",
			Help: "Current number of inflight HTTP requests",
		}),
		RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations",
			Buckets: prometheus.DefBuckets,
		}, []string{method, endpoint}),
	}
}
