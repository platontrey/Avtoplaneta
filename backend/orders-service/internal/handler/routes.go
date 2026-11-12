package handler

import (
	"github.com/gin-gonic/gin"

	"avtoplaneta/pkg/metrics"
)

// SetupRoutes настраивает маршруты для приложения с dependency injection
func SetupRoutes(r *gin.Engine, handler *Handler) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Add metrics middleware via shared pkg/metrics (с защитой от Cardinality Explosion через c.FullPath())
	r.Use(metrics.GinMiddleware("orders-service"))

	// Add /metrics endpoint
	r.GET("/metrics", metrics.Handler())

	// Применение middleware аутентификации ко всем бизнес-маршрутам
	authorized := r.Group("/")
	authorized.Use(AuthMiddleware())
	{
		// Маршруты администраторов (Legacy)
		admin := authorized.Group("/admin")
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

		// Обычные пользовательские маршруты (Legacy)
		authorized.POST("/orders", handler.CreateOrderHandler)
		authorized.GET("/orders", handler.GetOrdersHandler)
		authorized.GET("/orders/completed", handler.GetCompletedOrdersHandler)
		authorized.PUT("/orders/:id", handler.UpdateOrderDetailsHandler)
		authorized.PATCH("/orders/:id", handler.UpdateOrderDetailsHandler)
		authorized.PUT("/orders/:id/status", handler.UpdateOrderStatusHandler)
		authorized.PUT("/orders/:id/complete", handler.CompleteOrderHandler)
		authorized.DELETE("/orders/:id", handler.DeleteOrderHandler)
		authorized.POST("/orders/:id/items", handler.AddOrderItemHandler)
		authorized.PUT("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
		authorized.PATCH("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
		authorized.DELETE("/orders/:id/items/:itemId", handler.DeleteOrderItemHandler)
		authorized.GET("/monthly-sales", handler.GetMonthlySalesHandler)

		// Клиенты (Legacy)
		authorized.GET("/orders/customers", handler.ListCustomersHandler)
		authorized.POST("/orders/customers", handler.CreateCustomerHandler)
		authorized.GET("/orders/customers/:id", handler.GetCustomerHandler)
		authorized.PATCH("/orders/customers/:id", handler.UpdateCustomerHandler)
		authorized.PUT("/orders/customers/:id", handler.UpdateCustomerHandler)
		authorized.DELETE("/orders/customers/:id", handler.DeleteCustomerHandler)
		authorized.GET("/customers", handler.ListCustomersHandler)
		authorized.POST("/customers", handler.CreateCustomerHandler)
		authorized.GET("/customers/:id", handler.GetCustomerHandler)
		authorized.PATCH("/customers/:id", handler.UpdateCustomerHandler)
		authorized.PUT("/customers/:id", handler.UpdateCustomerHandler)
		authorized.DELETE("/customers/:id", handler.DeleteCustomerHandler)

		// Основные v1 маршруты
		v1 := authorized.Group("/api/v1")
		{
			v1Admin := v1.Group("/admin")
			v1Admin.Use(AdminMiddleware())
			{
				v1Admin.GET("/orders", handler.GetOrdersHandler)
				v1Admin.GET("/orders/completed", handler.GetCompletedOrdersHandler)
				v1Admin.PUT("/orders/:id", handler.UpdateOrderDetailsHandler)
				v1Admin.PATCH("/orders/:id", handler.UpdateOrderDetailsHandler)
				v1Admin.PUT("/orders/:id/status", handler.UpdateOrderStatusHandler)
				v1Admin.PUT("/orders/:id/complete", handler.CompleteOrderHandler)
				v1Admin.DELETE("/orders/:id", handler.DeleteOrderHandler)
				v1Admin.PUT("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
				v1Admin.PATCH("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
				v1Admin.DELETE("/orders/:id/items/:itemId", handler.DeleteOrderItemHandler)
			}

			v1.POST("/orders", handler.CreateOrderHandler)
			v1.GET("/orders", handler.GetOrdersHandler)
			v1.GET("/orders/completed", handler.GetCompletedOrdersHandler)
			v1.PUT("/orders/:id", handler.UpdateOrderDetailsHandler)
			v1.PATCH("/orders/:id", handler.UpdateOrderDetailsHandler)
			v1.PUT("/orders/:id/status", handler.UpdateOrderStatusHandler)
			v1.PUT("/orders/:id/complete", handler.CompleteOrderHandler)
			v1.DELETE("/orders/:id", handler.DeleteOrderHandler)
			v1.POST("/orders/:id/items", handler.AddOrderItemHandler)
			v1.PUT("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
			v1.PATCH("/orders/:id/items/:itemId", handler.UpdateOrderItemHandler)
			v1.DELETE("/orders/:id/items/:itemId", handler.DeleteOrderItemHandler)
			v1.GET("/orders/monthly-sales", handler.GetMonthlySalesHandler)

			// Клиенты (v1)
			v1.GET("/orders/customers", handler.ListCustomersHandler)
			v1.POST("/orders/customers", handler.CreateCustomerHandler)
			v1.GET("/orders/customers/:id", handler.GetCustomerHandler)
			v1.PATCH("/orders/customers/:id", handler.UpdateCustomerHandler)
			v1.PUT("/orders/customers/:id", handler.UpdateCustomerHandler)
			v1.DELETE("/orders/customers/:id", handler.DeleteCustomerHandler)

			v1.GET("/customers", handler.ListCustomersHandler)
			v1.POST("/customers", handler.CreateCustomerHandler)
			v1.GET("/customers/:id", handler.GetCustomerHandler)
			v1.PATCH("/customers/:id", handler.UpdateCustomerHandler)
			v1.PUT("/customers/:id", handler.UpdateCustomerHandler)
			v1.DELETE("/customers/:id", handler.DeleteCustomerHandler)
		}
	}
}
