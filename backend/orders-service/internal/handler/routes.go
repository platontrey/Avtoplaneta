package handler

import (
	"github.com/gin-gonic/gin"

	"avtoplaneta/pkg/metrics"
)

// SetupRoutes настраивает маршруты для приложения с dependency injection
func SetupRoutes(r *gin.Engine, handler *Handler) {
	// Регистрация health-проверок (/health, /health/live, /health/ready, /live, /ready)
	handler.HealthChecker().RegisterRoutes(r)

	// Add metrics middleware via shared pkg/metrics (с защитой от Cardinality Explosion через c.FullPath())
	r.Use(metrics.GinMiddleware("orders-service"))

	// Add /metrics endpoint
	r.GET("/metrics", metrics.Handler())

	// Применение middleware аутентификации ко всем бизнес-маршрутам
	authorized := r.Group("/")
	authorized.Use(AuthMiddleware())

	registerRoutesForGroup := func(g *gin.RouterGroup) {
		admin := g.Group("/admin")
		admin.Use(AdminMiddleware())
		{
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
		}

		g.POST("/orders", handler.CreateOrderHandler)
		g.GET("/orders", handler.GetOrdersHandler)
		g.GET("/orders/completed", handler.GetCompletedOrdersHandler)
		g.PUT("/orders/:id", handler.UpdateOrderDetailsHandler)
		g.PATCH("/orders/:id", handler.UpdateOrderDetailsHandler)
		g.PUT("/orders/:id/status", handler.UpdateOrderStatusHandler)
		g.PUT("/orders/:id/complete", handler.CompleteOrderHandler)
		g.DELETE("/orders/:id", handler.DeleteOrderHandler)
		g.POST("/orders/:id/items", handler.AddOrderItemHandler)
		g.PUT("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
		g.PATCH("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
		g.DELETE("/orders/:id/items/:itemId", handler.DeleteOrderItemHandler)
		g.GET("/monthly-sales", handler.GetMonthlySalesHandler)

		// Клиенты
		g.GET("/orders/customers", handler.ListCustomersHandler)
		g.POST("/orders/customers", handler.CreateCustomerHandler)
		g.GET("/orders/customers/:id", handler.GetCustomerHandler)
		g.PATCH("/orders/customers/:id", handler.UpdateCustomerHandler)
		g.PUT("/orders/customers/:id", handler.UpdateCustomerHandler)
		g.DELETE("/orders/customers/:id", handler.DeleteCustomerHandler)
		g.GET("/customers", handler.ListCustomersHandler)
		g.POST("/customers", handler.CreateCustomerHandler)
		g.GET("/customers/:id", handler.GetCustomerHandler)
		g.PATCH("/customers/:id", handler.UpdateCustomerHandler)
		g.PUT("/customers/:id", handler.UpdateCustomerHandler)
		g.DELETE("/customers/:id", handler.DeleteCustomerHandler)
	}

	// Legacy маршруты
	registerRoutesForGroup(authorized)

	// Основные v1 маршруты (/api/v1)
	registerRoutesForGroup(authorized.Group("/api/v1"))
}
