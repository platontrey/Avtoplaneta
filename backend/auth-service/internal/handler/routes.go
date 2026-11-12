package handler

import (
	"github.com/gin-gonic/gin"

	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/metrics"
)

// SetupRoutes настраивает маршруты для приложения с dependency injection
func SetupRoutes(r *gin.Engine, h *Handler) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Add metrics middleware via shared pkg/metrics (с защитой от Cardinality Explosion через c.FullPath())
	r.Use(metrics.GinMiddleware("auth-service"))

	// Add /metrics endpoint
	r.GET("/metrics", metrics.Handler())

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
	r.GET("/auth/verify-admin", h.AuthMiddleware, h.RequireRole(constants.RoleAdmin), h.VerifyAuthHandler)

	// Список пользователей для мессенджера и операторов
	r.GET("/api/users", h.AuthMiddleware, h.RequireMinRole(constants.RoleOperator), h.GetUsersHandler)
	r.GET("/api/users/me", h.AuthMiddleware, h.GetCurrentUserHandler)

	// Управление личным профилем текущего пользователя
	r.PUT("/auth/profile", h.AuthMiddleware, h.UpdateProfileHandler)
	r.PUT("/auth/me", h.AuthMiddleware, h.UpdateProfileHandler)
	r.POST("/auth/change-password", h.AuthMiddleware, h.ChangePasswordHandler)
	r.POST("/auth/avatar", h.AuthMiddleware, h.UploadAvatarHandler)
	r.DELETE("/auth/avatar", h.AuthMiddleware, h.DeleteAvatarHandler)

	// Зеркалирование для API клиентов /api/users/me
	r.PUT("/api/users/me", h.AuthMiddleware, h.UpdateProfileHandler)
	r.POST("/api/users/me/password", h.ChangePasswordHandler)
	r.POST("/api/users/me/avatar", h.UploadAvatarHandler)
	r.DELETE("/api/users/me/avatar", h.DeleteAvatarHandler)

	// Раздача статических аватаров
	r.Static("/auth/avatars", "./uploads/avatars")

	// Автообновление мобильного приложения
	r.GET("/api/app/version", h.GetAppVersionHandler)
	r.HEAD("/api/app/version", h.GetAppVersionHandler)
	r.GET("/api/app/download", h.DownloadAppHandler)
	r.HEAD("/api/app/download", h.DownloadAppHandler)

	// Маршруты панели администратора (канонические /api/v1/admin, /api/admin и устаревшие /admin)
	registerAdminRoutes := func(prefix string) {
		admin := r.Group(prefix)
		admin.Use(h.AuthMiddleware)
		{
			// Управление пользователями - требуется роль admin
			admin.POST("/users", h.RequireRole(constants.RoleAdmin), h.CreateUserHandler)
			admin.GET("/users", h.RequireMinRole(constants.RoleManager), h.GetUsersHandler)
			admin.PUT("/users/:id", h.RequireRole(constants.RoleAdmin), h.UpdateUserHandler)
			admin.DELETE("/users/:id", h.RequireRole(constants.RoleAdmin), h.DeleteUserHandler)

			// Управление сервером - требуется роль admin
			admin.GET("/status", h.RequireRole(constants.RoleAdmin), h.GetServerStatusHandler)
			admin.GET("/logs", h.RequireRole(constants.RoleAdmin), h.GetServerLogsHandler)

			// Логи активности пользователей - требуется роль admin
			admin.GET("/user-activity-logs", h.RequireRole(constants.RoleAdmin), h.GetUserActivityLogsHandler)
			admin.POST("/user-activity-logs", h.RequireRole(constants.RoleAdmin), h.LogUserActivityHandler)
		}
	}

	registerAdminRoutes("/api/v1/admin")
	registerAdminRoutes("/api/admin")
	registerAdminRoutes("/admin")

	// Внутренний endpoint для логирования активности без аутентификации (только для внутренних сервисов)
	r.POST("/internal/log-activity", h.InternalLogUserActivityHandler)

	// Внутренний endpoint для получения пользователей без аутентификации (только для внутренних сервисов)
	r.GET("/internal/users", h.InternalGetUsersHandler)
}
