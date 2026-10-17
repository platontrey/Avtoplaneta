package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registerOnce sync.Once

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество входящих HTTP-запросов",
		},
		[]string{"service", "method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Длительность обработки HTTP-запросов в секундах",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "path"},
	)

	httpRequestsErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_errors_total",
			Help: "Общее количество ошибок HTTP-запросов (4xx, 5xx)",
		},
		[]string{"service", "method", "path", "status"},
	)
)

func init() {
	registerMetrics(prometheus.DefaultRegisterer)
}

func registerMetrics(reg prometheus.Registerer) {
	registerOnce.Do(func() {
		reg.MustRegister(
			httpRequestsTotal,
			httpRequestDuration,
			httpRequestsErrorsTotal,
		)
	})
}

// GinMiddleware возвращает middleware для Gin Engine с автоматической фиксацией RPS, Latency и ошибок
func GinMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		httpRequestsTotal.WithLabelValues(serviceName, method, path, status).Inc()
		httpRequestDuration.WithLabelValues(serviceName, method, path).Observe(duration)

		if c.Writer.Status() >= 400 {
			httpRequestsErrorsTotal.WithLabelValues(serviceName, method, path, status).Inc()
		}
	}
}

// Handler возвращает стандартный HTTP-обработчик для эндпоинта /metrics
func Handler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// OutboxMetrics хранит метрики Transactional Outbox для микросервиса
type OutboxMetrics struct {
	serviceName        string
	EventsProcessed    *prometheus.CounterVec
	EventsFailed       *prometheus.CounterVec
	ProcessingDuration *prometheus.HistogramVec
}

var (
	outboxOnce      sync.Once
	eventsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "outbox_events_processed_total",
			Help: "Общее количество успешно обработанных событий Outbox",
		},
		[]string{"service", "event_type"},
	)
	eventsFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "outbox_events_failed_total",
			Help: "Общее количество сбоев при обработке событий Outbox",
		},
		[]string{"service", "event_type"},
	)
	processingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "outbox_processing_duration_seconds",
			Help:    "Длительность обработки событий Outbox в секундах",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"service", "event_type"},
	)
)

// NewOutboxMetrics создает и регистрирует метрики Outbox для конкретного сервиса
func NewOutboxMetrics(serviceName string) *OutboxMetrics {
	outboxOnce.Do(func() {
		prometheus.MustRegister(
			eventsProcessed,
			eventsFailed,
			processingDuration,
		)
	})

	return &OutboxMetrics{
		serviceName:        serviceName,
		EventsProcessed:    eventsProcessed,
		EventsFailed:       eventsFailed,
		ProcessingDuration: processingDuration,
	}
}

// IncProcessed увеличивает счетчик успешно обработанных событий
func (m *OutboxMetrics) IncProcessed(eventType string) {
	m.EventsProcessed.WithLabelValues(m.serviceName, eventType).Inc()
}

// IncFailed увеличивает счетчик ошибок обработки событий
func (m *OutboxMetrics) IncFailed(eventType string) {
	m.EventsFailed.WithLabelValues(m.serviceName, eventType).Inc()
}

// ObserveDuration фиксирует длительность обработки события
func (m *OutboxMetrics) ObserveDuration(eventType string, duration time.Duration) {
	m.ProcessingDuration.WithLabelValues(m.serviceName, eventType).Observe(duration.Seconds())
}
