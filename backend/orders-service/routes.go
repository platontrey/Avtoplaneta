package main

import "github.com/gin-gonic/gin"

// SetupRoutes настраивает маршруты для приложения с dependency injection
func SetupRoutes(r *gin.Engine, handler *Handler) {
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
}