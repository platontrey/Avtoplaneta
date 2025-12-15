package main

import "github.com/gorilla/mux"

// SetupRoutes настраивает маршруты для приложения с dependency injection
func SetupRoutes(r *mux.Router, handler *Handler) {
	// Применение middleware аутентификации ко всем маршрутам
	r.Use(authMiddleware)

	// Маршруты только для администраторов
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(adminMiddleware)
	admin.HandleFunc("/orders", handler.GetOrdersHandler).Methods("GET", "OPTIONS")
	admin.HandleFunc("/orders/{id:[0-9]+}/status", handler.UpdateOrderStatusHandler).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/orders/{id:[0-9]+}/complete", handler.CompleteOrderHandler).Methods("PUT", "OPTIONS")
	admin.HandleFunc("/orders/{id:[0-9]+}", handler.DeleteOrderHandler).Methods("DELETE", "OPTIONS")

	// Обычные пользовательские маршруты (для создания заказов)
	r.HandleFunc("/orders", handler.CreateOrderHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/orders", handler.GetOrdersHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/orders/{id:[0-9]+}/items", handler.AddOrderItemHandler).Methods("POST", "OPTIONS")
}