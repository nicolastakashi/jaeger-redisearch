package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var WritesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "jaeger_redis_inserts_total",
	Help: "Number of inserts in Redis.",
}, []string{"index"})

var WritesLantency = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "jaeger_redis_inserts_latency",
	Help: "Latency of inserts in Redis.",
}, []string{"index", "status"})

var ReadsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "jaeger_redis_read_total",
	Help: "Number of read in Redis.",
}, []string{"index", "operation"})

var ReadLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "jaeger_redis_read_latency",
	Help: "Latency of read in Redis.",
}, []string{"index", "status", "operation"})
