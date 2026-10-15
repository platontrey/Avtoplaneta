package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	OutboxEventsProcessedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_outbox_events_processed_total",
			Help: "Total number of processed orders outbox events",
		},
		[]string{"event_type"},
	)
	OutboxEventsFailedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orders_outbox_events_failed_total",
			Help: "Total number of failed orders outbox events",
		},
		[]string{"event_type"},
	)
	OutboxProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "orders_outbox_processing_duration_seconds",
			Help:    "Duration of orders outbox event processing in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"event_type"},
	)
)

func init() {
	prometheus.MustRegister(
		OutboxEventsProcessedTotal,
		OutboxEventsFailedTotal,
		OutboxProcessingDuration,
	)
}
