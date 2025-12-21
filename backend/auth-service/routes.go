package main

import "github.com/gin-gonic/gin"

// SetupRoutes настраивает маршруты для приложения с dependency injection
func SetupRoutes(r *gin.Engine, handler *Handler) {
	// Основные маршруты аутентификации
	r.GET("/auth/google", handler.GoogleAuthHandler)
	r.GET("/auth/google/callback", handler.GoogleAuthCallbackHandler)
	r.POST("/auth/login", handler.UserLoginHandler)
	r.POST("/auth/logout", handler.LogoutHandler)
	r.GET("/auth/me", handler.GetCurrentUserHandler)
	r.GET("/auth/csrf-token", handler.GetCSRFTokenHandler)

	// Маршруты панели администратора
	admin := r.Group("/admin")
	admin.Use(authMiddleware)
	{
		// Управление пользователями - требуется роль admin
		admin.POST("/users", requireRole("admin"), handler.CreateUserHandler)
		admin.GET("/users", requireMinRole("manager"), handler.GetUsersHandler)
		admin.PUT("/users/:id", requireRole("admin"), handler.UpdateUserHandler)
		admin.DELETE("/users/:id", requireRole("admin"), handler.DeleteUserHandler)

		// Управление сервером - требуется роль admin
		admin.GET("/status", requireRole("admin"), handler.GetServerStatusHandler)
		admin.GET("/logs", requireRole("admin"), handler.GetServerLogsHandler)

		// Логи активности пользователей - требуется роль admin
		admin.GET("/user-activity-logs", requireRole("admin"), handler.GetUserActivityLogsHandler)
		admin.POST("/user-activity-logs", requireRole("admin"), handler.LogUserActivityHandler)

		// Внутренний endpoint для логирования активности без аутентификации (только для внутренних сервисов)
		r.POST("/internal/log-activity", handler.InternalLogUserActivityHandler)

		// Внутренний endpoint для получения пользователей без аутентификации (только для внутренних сервисов)
		r.GET("/internal/users", handler.InternalGetUsersHandler)
	}
}