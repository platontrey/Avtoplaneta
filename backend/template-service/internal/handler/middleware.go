package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"avtoplaneta/pkg/authcontext"
	"avtoplaneta/pkg/metrics"
)

// CORSMiddleware стандартная настройка CORS для веб-клиента и шлюза Traefik
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-ID, X-User-Email, X-User-Name, X-User-Role")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
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

// AuthMiddleware извлекает пользователя из ForwardAuth заголовков Traefik
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authcontext.FromRequest(c.Request)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Set("user", user)
		c.Next()
	}
}
