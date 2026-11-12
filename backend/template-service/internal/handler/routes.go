package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRoutes настраивает все маршруты HTTP-сервера
func SetupRoutes(r *gin.Engine, h *Handler) {
	// Базовые middleware
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())
	r.Use(LoggingMiddleware())
	r.Use(MetricsMiddleware())

	// Регистрация health-проверок (/health, /health/live, /health/ready, /live, /ready)
	h.HealthChecker().RegisterRoutes(r)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Публичные или защищенные API-роуты
	v1 := r.Group("/api/v1")
	{
		items := v1.Group("/items")
		{
			items.GET("", h.ListItems)
			items.GET("/:id", h.GetItem)
			items.POST("", h.CreateItem)
			items.PUT("/:id", h.UpdateItem)
			items.DELETE("/:id", h.DeleteItem)
		}
	}
}
