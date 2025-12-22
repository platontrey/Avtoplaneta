package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Установка количества OS-тредов
	runtime.GOMAXPROCS(runtime.NumCPU())

	config := LoadConfig()
	InitDB(config)
	InitAuth(config)

	// Создание контекста с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Получен сигнал завершения, начинаем graceful shutdown...")
		cancel()
	}()

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

	// Создание HTTP сервера для graceful shutdown
	srv := &http.Server{
		Addr:    ":" + config.Port,
		Handler: r,
	}

	// Канал для ошибок сервера
	errChan := make(chan error, 1)

	// Запуск сервера в goroutine
	go func() {
		log.Println("Сервис заказов запущен на порту", config.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			errChan <- err
		}
	}()

	// Ожидание сигнала отмены или ошибки
	select {
	case <-ctx.Done():
		log.Println("Завершение работы сервиса заказов...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Сервер принудительно остановлен: %v", err)
		}
		log.Println("Сервис заказов остановлен")
	case err := <-errChan:
		log.Fatal("Ошибка сервера:", err)
	}
}
