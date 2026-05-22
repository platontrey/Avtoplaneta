package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "avtoplaneta_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	httpRequestsErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_http_requests_errors_total",
			Help: "Total number of HTTP request errors",
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

// SetupRoutes настраивает маршруты для приложения с dependency injection
func SetupRoutes(r *gin.Engine, handler *Handler) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Add metrics middleware
	r.Use(metricsMiddleware())

	// Add /metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Применение middleware аутентификации ко всем маршрутам
	r.Use(authMiddleware())

	// Маршруты только для администраторов
	admin := r.Group("/admin")
	admin.Use(adminMiddleware())
	admin.GET("/orders", handler.GetOrdersHandler)
	admin.PUT("/orders/:id/status", handler.UpdateOrderStatusHandler)
	admin.PUT("/orders/:id/complete", handler.CompleteOrderHandler)
	admin.DELETE("/orders/:id", handler.DeleteOrderHandler)

	// Обычные пользовательские маршруты (для создания заказов)
	r.POST("/orders", handler.CreateOrderHandler)
	r.GET("/orders", handler.GetOrdersHandler)
	r.POST("/orders/:id/items", handler.AddOrderItemHandler)
	r.GET("/monthly-sales", handler.GetMonthlySalesHandler)
}

// metricsMiddleware measures HTTP request latency and throughput
func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path

		// Proceed with the request
		c.Next()

		// Measure duration
		duration := time.Since(start).Seconds()

		// Get status code
		status := c.Writer.Status()

		// Record metrics
		httpRequestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)

		// Record errors if status >= 400
		if status >= 400 {
			httpRequestsErrorsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
		}
	}
}
