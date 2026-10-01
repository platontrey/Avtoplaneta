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

	// Маршруты администраторов (Legacy)
	admin := r.Group("/admin")
	admin.Use(adminMiddleware())
	admin.GET("/orders", handler.GetOrdersHandler)
	admin.GET("/orders/completed", handler.GetCompletedOrdersHandler)
	admin.PUT("/orders/:id", handler.UpdateOrderDetailsHandler)
	admin.PATCH("/orders/:id", handler.UpdateOrderDetailsHandler)
	admin.PUT("/orders/:id/status", handler.UpdateOrderStatusHandler)
	admin.PUT("/orders/:id/complete", handler.CompleteOrderHandler)
	admin.DELETE("/orders/:id", handler.DeleteOrderHandler)
	admin.PUT("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
	admin.PATCH("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
	admin.DELETE("/orders/:id/items/:itemId", handler.DeleteOrderItemHandler)

	// Обычные пользовательские маршруты (Legacy)
	r.POST("/orders", handler.CreateOrderHandler)
	r.GET("/orders", handler.GetOrdersHandler)
	r.GET("/orders/completed", handler.GetCompletedOrdersHandler)
	r.PUT("/orders/:id", handler.UpdateOrderDetailsHandler)
	r.PATCH("/orders/:id", handler.UpdateOrderDetailsHandler)
	r.PUT("/orders/:id/status", handler.UpdateOrderStatusHandler)
	r.PUT("/orders/:id/complete", handler.CompleteOrderHandler)
	r.DELETE("/orders/:id", handler.DeleteOrderHandler)
	r.POST("/orders/:id/items", handler.AddOrderItemHandler)
	r.PUT("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
	r.PATCH("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
	r.DELETE("/orders/:id/items/:itemId", handler.DeleteOrderItemHandler)
	r.GET("/monthly-sales", handler.GetMonthlySalesHandler)

	// Клиенты (Legacy)
	r.GET("/orders/customers", handler.ListCustomersHandler)
	r.POST("/orders/customers", handler.CreateCustomerHandler)
	r.GET("/orders/customers/:id", handler.GetCustomerHandler)
	r.PATCH("/orders/customers/:id", handler.UpdateCustomerHandler)
	r.PUT("/orders/customers/:id", handler.UpdateCustomerHandler)
	r.DELETE("/orders/customers/:id", handler.DeleteCustomerHandler)
	r.GET("/customers", handler.ListCustomersHandler)
	r.POST("/customers", handler.CreateCustomerHandler)
	r.GET("/customers/:id", handler.GetCustomerHandler)
	r.PATCH("/customers/:id", handler.UpdateCustomerHandler)
	r.PUT("/customers/:id", handler.UpdateCustomerHandler)
	r.DELETE("/customers/:id", handler.DeleteCustomerHandler)

	// Основные v1 маршруты
	v1 := r.Group("/api/v1")
	{
		v1Admin := v1.Group("/admin")
		v1Admin.Use(adminMiddleware())
		v1Admin.GET("/orders", handler.GetOrdersHandler)
		v1Admin.GET("/orders/completed", handler.GetCompletedOrdersHandler)
		v1Admin.PUT("/orders/:id", handler.UpdateOrderDetailsHandler)
		v1Admin.PATCH("/orders/:id", handler.UpdateOrderDetailsHandler)
		v1Admin.PUT("/orders/:id/status", handler.UpdateOrderStatusHandler)
		v1Admin.PUT("/orders/:id/complete", handler.CompleteOrderHandler)
		v1Admin.DELETE("/orders/:id", handler.DeleteOrderHandler)
		v1Admin.PUT("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
		v1Admin.PATCH("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
		v1Admin.DELETE("/orders/:id/items/:itemId", handler.DeleteOrderItemHandler)

		v1.POST("/orders", handler.CreateOrderHandler)
		v1.GET("/orders", handler.GetOrdersHandler)
		v1.GET("/orders/completed", handler.GetCompletedOrdersHandler)
		v1.PUT("/orders/:id", handler.UpdateOrderDetailsHandler)
		v1.PATCH("/orders/:id", handler.UpdateOrderDetailsHandler)
		v1.PUT("/orders/:id/status", handler.UpdateOrderStatusHandler)
		v1.PUT("/orders/:id/complete", handler.CompleteOrderHandler)
		v1.DELETE("/orders/:id", handler.DeleteOrderHandler)
		v1.POST("/orders/:id/items", handler.AddOrderItemHandler)
		v1.PUT("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
		v1.PATCH("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
		v1.DELETE("/orders/:id/items/:itemId", handler.DeleteOrderItemHandler)
		v1.GET("/orders/monthly-sales", handler.GetMonthlySalesHandler)

		// Клиенты (v1)
		v1.GET("/orders/customers", handler.ListCustomersHandler)
		v1.POST("/orders/customers", handler.CreateCustomerHandler)
		v1.GET("/orders/customers/:id", handler.GetCustomerHandler)
		v1.PATCH("/orders/customers/:id", handler.UpdateCustomerHandler)
		v1.PUT("/orders/customers/:id", handler.UpdateCustomerHandler)
		v1.DELETE("/orders/customers/:id", handler.DeleteCustomerHandler)

		v1.GET("/customers", handler.ListCustomersHandler)
		v1.POST("/customers", handler.CreateCustomerHandler)
		v1.GET("/customers/:id", handler.GetCustomerHandler)
		v1.PATCH("/customers/:id", handler.UpdateCustomerHandler)
		v1.PUT("/customers/:id", handler.UpdateCustomerHandler)
		v1.DELETE("/customers/:id", handler.DeleteCustomerHandler)
	}
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
