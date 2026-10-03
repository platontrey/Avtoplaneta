package main

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)
var (
	dbErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "service"},
	)
	businessOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_business_operations_total",
			Help: "Total number of business operations",
		},
		[]string{"operation", "service", "status"},
	)
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

func RecordDBError(operation, service string) {
	dbErrorsTotal.WithLabelValues(operation, service).Inc()
}
func RecordBusinessOperation(operation, service, status string) {
	businessOperationsTotal.WithLabelValues(operation, service, status).Inc()
}

func init() {
	// Register metrics with Prometheus (avoid registering multiple times if init runs twice in tests)
	_ = prometheus.Register(dbErrorsTotal)
	_ = prometheus.Register(businessOperationsTotal)
	_ = prometheus.Register(httpRequestsTotal)
	_ = prometheus.Register(httpRequestDuration)
	_ = prometheus.Register(httpRequestsErrorsTotal)
}

func setupRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.Use(metricsMiddleware())

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Основные v1 маршруты
	registerMessagingEndpoints(r.Group("/api/v1/messaging"))

	// Legacy-алиасы для обратной совместимости
	registerMessagingEndpoints(r.Group("/api/messaging"))
}

func registerMessagingEndpoints(api *gin.RouterGroup) {
	// WebSocket
	api.GET("/ws", ServeWebSocket)

	// Conversations
	api.GET("/conversations", getConversations)
	api.POST("/conversations", createConversation)
	api.GET("/conversations/:id", getConversation)
	api.PUT("/conversations/:id", updateConversation)
	api.DELETE("/conversations/:id", deleteConversation)
	api.DELETE("/conversations/:id/participants/:userId", removeParticipant)

	// Messages
	api.GET("/conversations/:id/messages", getMessages)
	api.POST("/conversations/:id/messages", sendMessage)
	api.POST("/conversations/:id/messages/voice", sendVoiceMessage)
	api.DELETE("/messages/:id", deleteMessage)
	api.PUT("/messages/:id/read", markMessageRead)

	// Reactions
	api.POST("/messages/:id/reactions", addReaction)
	api.DELETE("/messages/:id/reactions/:reactionId", removeReaction)

	// Notifications
	api.GET("/notifications", getNotifications)
	api.PUT("/notifications/:id/read", markNotificationRead)
	api.PUT("/notifications/read-all", markAllNotificationsRead)

	// User status
	api.GET("/users/status", getUserStatuses)
	api.PUT("/users/status", updateUserStatus)

	// Users
	api.GET("/users", getUsers)

	// Search
	api.GET("/search", searchMessages)

	// Drom
	api.GET("/drom/dialogs", getDromDialogs)
	api.GET("/drom/messages", getDromMessages)
	api.POST("/drom/messages", sendDromMessage)
}

func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}

		httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)

		if c.Writer.Status() >= 400 {
			httpRequestsErrorsTotal.WithLabelValues(method, endpoint, status).Inc()
		}
	}
}
