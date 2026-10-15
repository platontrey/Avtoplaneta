package metrics

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_auth_http_requests_total",
			Help: "Total number of HTTP requests in auth service",
		},
		[]string{"method", "endpoint", "status"},
	)
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "avtoplaneta_auth_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds in auth service",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	HTTPRequestsErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_auth_http_requests_errors_total",
			Help: "Total number of HTTP request errors in auth service",
		},
		[]string{"method", "endpoint", "status"},
	)
	OutboxEventsProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_outbox_events_processed_total",
			Help: "Total number of processed auth outbox events",
		},
		[]string{"event_type"},
	)
	OutboxEventsFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_outbox_events_failed_total",
			Help: "Total number of failed auth outbox events",
		},
		[]string{"event_type"},
	)
	OutboxProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_outbox_processing_duration_seconds",
			Help:    "Duration of auth outbox event processing in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"event_type"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		HTTPRequestsErrorsTotal,
		OutboxEventsProcessedTotal,
		OutboxEventsFailedTotal,
		OutboxProcessingDuration,
	)
}

// MetricsMiddleware собирает метрики Prometheus для HTTP запросов
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		HTTPRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
		HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

		if status >= 400 {
			HTTPRequestsErrorsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
		}
	}
}
