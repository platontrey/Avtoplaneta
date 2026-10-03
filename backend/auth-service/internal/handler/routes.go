package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"auth-service/internal/metrics"
)

// SetupRoutes настраивает маршруты для приложения с dependency injection
func SetupRoutes(r *gin.Engine, h *Handler) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Add metrics middleware
	r.Use(metrics.MetricsMiddleware())

	// Add /metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// CORS middleware
	r.Use(h.CORSMiddleware())

	// Основные маршруты аутентификации
	r.GET("/auth/google", h.GoogleAuthHandler)
	r.GET("/auth/google/callback", h.GoogleAuthCallbackHandler)
	r.POST("/auth/google/mobile", h.GoogleMobileAuthHandler)
	r.POST("/auth/login", h.UserLoginHandler)
	r.POST("/auth/logout", h.LogoutHandler)
	r.GET("/auth/me", h.GetCurrentUserHandler)
	r.GET("/auth/csrf-token", h.GetCSRFTokenHandler)
	r.POST("/auth/refresh", h.RefreshTokenHandler)

	// Traefik ForwardAuth эндпоинты
	r.GET("/auth/verify", h.AuthMiddleware, h.VerifyAuthHandler)
	r.GET("/auth/verify-admin", h.AuthMiddleware, h.RequireRole("admin"), h.VerifyAuthHandler)

	// Список пользователей для мессенджера и операторов
	r.GET("/api/users", h.AuthMiddleware, h.RequireMinRole("operator"), h.GetUsersHandler)
	r.GET("/api/users/me", h.AuthMiddleware, h.GetCurrentUserHandler)

	// Автообновление мобильного приложения
	r.GET("/api/app/version", h.GetAppVersionHandler)
	r.GET("/api/app/download", h.DownloadAppHandler)

	// Маршруты панели администратора
	admin := r.Group("/admin")
	admin.Use(h.AuthMiddleware)
	{
		// Управление пользователями - требуется роль admin
		admin.POST("/users", h.RequireRole("admin"), h.CreateUserHandler)
		admin.GET("/users", h.RequireMinRole("manager"), h.GetUsersHandler)
		admin.PUT("/users/:id", h.RequireRole("admin"), h.UpdateUserHandler)
		admin.DELETE("/users/:id", h.RequireRole("admin"), h.DeleteUserHandler)

		// Управление сервером - требуется роль admin
		admin.GET("/status", h.RequireRole("admin"), h.GetServerStatusHandler)
		admin.GET("/logs", h.RequireRole("admin"), h.GetServerLogsHandler)

		// Логи активности пользователей - требуется роль admin
		admin.GET("/user-activity-logs", h.RequireRole("admin"), h.GetUserActivityLogsHandler)
		admin.POST("/user-activity-logs", h.RequireRole("admin"), h.LogUserActivityHandler)

		// Внутренний endpoint для логирования активности без аутентификации (только для внутренних сервисов)
		r.POST("/internal/log-activity", h.InternalLogUserActivityHandler)

		// Внутренний endpoint для получения пользователей без аутентификации (только для внутренних сервисов)
		r.GET("/internal/users", h.InternalGetUsersHandler)
	}
}
