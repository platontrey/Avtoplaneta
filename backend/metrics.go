package main

import (
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// Prometheus метрики
var (
	// Счетчики HTTP запросов
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Гистограмма длительности HTTP запросов
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "avtoplaneta_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Счетчик активных подключений
	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "avtoplaneta_active_connections",
			Help: "Number of active connections",
		},
	)

	// Счетчик ошибок базы данных
	dbErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "service"},
	)

	// Счетчик ошибок Elasticsearch
	esErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "avtoplaneta_elasticsearch_errors_total",
			Help: "Total number of Elasticsearch errors",
		},
	)

	// Гистограмма размера ответов
	responseSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "avtoplaneta_response_size_bytes",
			Help:    "Response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 2, 10),
		},
		[]string{"endpoint"},
	)

	// Счетчик бизнес-операций
	businessOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_business_operations_total",
			Help: "Total number of business operations",
		},
		[]string{"operation", "service", "status"},
	)

	// Gauge для количества активных goroutines
	activeGoroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "avtoplaneta_active_goroutines",
			Help: "Number of active goroutines",
		},
	)
)

// InitMetrics инициализирует Prometheus метрики
func InitMetrics() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		activeConnections,
		dbErrorsTotal,
		esErrorsTotal,
		responseSizeBytes,
		businessOperationsTotal,
		activeGoroutines,
	)
}

// MetricsMiddleware middleware для сбора метрик HTTP запросов
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		activeConnections.Inc()
		UpdateGoroutinesMetric() // Обновляем метрику goroutines

		// Определяем endpoint (убираем параметры пути)
		endpoint := c.Request.URL.Path
		if len(c.Params) > 0 {
			// Для путей с параметрами используем шаблон
			for _, param := range c.Params {
				endpoint = strings.Replace(endpoint, "/"+param.Value, "/:"+param.Key, 1)
			}
		}

		c.Next()

		// Собираем метрики после обработки запроса
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, endpoint).Observe(duration)
		responseSizeBytes.WithLabelValues(endpoint).Observe(float64(c.Writer.Size()))

		activeConnections.Dec()

		// Логируем медленные запросы
		if duration > 1.0 {
			logrus.WithFields(logrus.Fields{
				"method":   c.Request.Method,
				"path":     c.Request.URL.Path,
				"duration": duration,
				"status":   status,
			}).Warn("Slow request detected")
		}
	}
}


// MetricsHandler возвращает HTTP handler для Prometheus метрик
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// HealthCheckHandler проверка здоровья сервиса
func HealthCheckHandler(c *gin.Context) {
	health := map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
		"version": "1.0.0",
	}

	// Здесь можно добавить проверки подключений к БД, Redis и т.д.

	c.JSON(200, health)
}

// ReadinessHandler проверка готовности сервиса
func ReadinessHandler(c *gin.Context) {
	// Проверяем готовность зависимостей
	readiness := map[string]interface{}{
		"status": "ready",
		"checks": map[string]bool{
			"database": true, // Здесь можно добавить реальные проверки
			"cache":    true,
		},
	}

	c.JSON(200, readiness)
}

// RecordDBError записывает ошибку базы данных
func RecordDBError(operation, service string) {
	dbErrorsTotal.WithLabelValues(operation, service).Inc()
}

// RecordESError записывает ошибку Elasticsearch
func RecordESError() {
	esErrorsTotal.Inc()
}

// RecordBusinessOperation записывает бизнес-операцию
func RecordBusinessOperation(operation, service, status string) {
	businessOperationsTotal.WithLabelValues(operation, service, status).Inc()
}

// UpdateGoroutinesMetric обновляет метрику количества активных goroutines
func UpdateGoroutinesMetric() {
	activeGoroutines.Set(float64(runtime.NumGoroutine()))
}