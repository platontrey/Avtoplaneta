package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	config := LoadConfig()
	InitDB(config)
	InitAuth(config)

	// Создаем зависимости
	orderRepo := NewOrderRepository(db)
	partRepo := NewPartRepositoryForOrders(db)
	ordersService := NewOrdersService(orderRepo, partRepo)
	handler := NewHandler(ordersService)

	r := mux.NewRouter()

	// CORS middleware
	r.Use(CORSMiddleware)

	// Setup routes с dependency injection
	SetupRoutes(r, handler)

	log.Println("Сервис заказов запущен на порту", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, r))
}
