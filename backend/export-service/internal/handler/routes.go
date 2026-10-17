package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"export-service/internal/service"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_export_http_requests_total",
			Help: "Total number of HTTP requests in export-service",
		},
		[]string{"method", "endpoint", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "avtoplaneta_export_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds in export-service",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	httpRequestsErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_export_http_requests_errors_total",
			Help: "Total number of HTTP request errors in export-service",
		},
		[]string{"method", "endpoint", "status"},
	)
)

func init() {
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		httpRequestsErrorsTotal,
	)
}

func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		httpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)

		if status >= 400 {
			httpRequestsErrorsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
		}
	}
}

// SetupRoutes конфигурирует HTTP-маршруты сервиса экспорта
func SetupRoutes(router *gin.Engine, h *Handler) {
	router.Use(metricsMiddleware())

	router.GET("/health", h.HealthCheck)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/api/export/xml", h.ExportXMLPriceList)
	router.POST("/api/export/drom", h.SendPriceListToDrom)
	router.GET("/api/export/status", h.GetExportStatus)

	router.GET(service.PriceListURL, h.ServePriceList)
	router.HEAD(service.PriceListURL, h.ServePriceList)
}
