package handler

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

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
		slog.InfoContext(c.Request.Context(), "HTTP request",
			slog.Int("status", c.Writer.Status()),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("ip", c.ClientIP()),
			slog.Duration("latency", latency),
		)
	}
}

// AuthMiddleware извлекает пользователя из ForwardAuth заголовков Traefik через ginmw.RequireAuth
func AuthMiddleware() gin.HandlerFunc {
	return ginmw.RequireAuth()
}
