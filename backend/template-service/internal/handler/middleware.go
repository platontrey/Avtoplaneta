package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"avtoplaneta/pkg/ginmw"
	"avtoplaneta/pkg/metrics"
)

// CORSMiddleware стандартная настройка CORS для веб-клиента и шлюза Traefik через ginmw.CORS
func CORSMiddleware() gin.HandlerFunc {
	return ginmw.CORS()
}

// MetricsMiddleware собирает Prometheus метрики по каждому эндпоинту через shared pkg/metrics
func MetricsMiddleware() gin.HandlerFunc {
	return metrics.GinMiddleware("template-service")
}

// LoggingMiddleware логирует каждый запрос с задержкой и статусом
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		logrus.WithFields(logrus.Fields{
			"status":  c.Writer.Status(),
			"method":  c.Request.Method,
			"path":    c.Request.URL.Path,
			"ip":      c.ClientIP(),
			"latency": latency.String(),
		}).Info("HTTP request")
	}
}

// AuthMiddleware извлекает пользователя из ForwardAuth заголовков Traefik через ginmw.RequireAuth
func AuthMiddleware() gin.HandlerFunc {
	return ginmw.RequireAuth()
}
