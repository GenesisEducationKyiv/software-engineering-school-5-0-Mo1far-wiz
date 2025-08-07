package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_requests_total",
			Help: "Total number of processed requests",
		},
		[]string{"service", "status"},
	)
	RequestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_request_latency_seconds",
			Help:    "Request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service"},
	)
	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_errors_total",
			Help: "Total error count by service",
		},
		[]string{"service", "error_type"},
	)
	CacheOpsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_cache_ops_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "status"},
	)
	CacheLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_cache_latency_seconds",
			Help:    "Cache operation latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
	CacheSizeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_cache_size",
			Help: "Counts of various cacheable items",
		},
		[]string{"type", "frequency"},
	)
	SubscriptionRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_subscription_requests_total",
			Help: "Total subscription service requests",
		},
		[]string{"method", "status"}, // method = subscribe|confirm|unsubscribe
	)
	SubscriptionLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_subscription_latency_seconds",
			Help:    "Latency of subscription service methods",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
	ActiveSubscriptions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_active_subscriptions",
			Help: "Number of active subscriptions by frequency",
		},
		[]string{"frequency"},
	)
)

func Init() {
	prometheus.MustRegister(
		RequestsTotal,
		RequestLatency,
		ErrorsTotal,
		CacheOpsTotal,
		CacheLatency,
		CacheSizeGauge,
		SubscriptionRequests,
		SubscriptionLatency,
		ActiveSubscriptions,
	)
}

func ObserveRequest(service string, f func() error) error {
	start := time.Now()
	err := f()
	duration := time.Since(start).Seconds()

	RequestsTotal.WithLabelValues(service, StatusLabel(err)).Inc()
	RequestLatency.WithLabelValues(service).Observe(duration)
	if err != nil {
		ErrorsTotal.WithLabelValues(service, TypeLabel(err)).Inc()
	}

	return err
}

func StatusLabel(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}
func TypeLabel(err error) string {
	if err == nil {
		return "none"
	}
	return fmt.Sprintf("%T", err)
}
