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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Установка количества OS-тредов для оптимизации под доступное количество ядер
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	InitDatabase()

	// Initialize Redis
	InitRedis()

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

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Static file serving for uploads
	r.Static("/uploads", "./uploads")

	// Set API_BASE_URL
	if os.Getenv("MESSAGING_SERVICE_URL") != "" {
		os.Setenv("API_BASE_URL", os.Getenv("MESSAGING_SERVICE_URL"))
	} else {
		os.Setenv("API_BASE_URL", "http://localhost:8084")
	}

	// Routes will be added here
	setupRoutes(r)

	// Создание HTTP сервера для graceful shutdown
	srv := &http.Server{
		Addr:    ":8084",
		Handler: r,
	}

	// Канал для ошибок сервера
	errChan := make(chan error, 1)

	// Запуск сервера в goroutine
	go func() {
		log.Println("Messaging service starting on port 8084")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			errChan <- err
		}
	}()

	// Ожидание сигнала отмены или ошибки
	select {
	case <-ctx.Done():
		log.Println("Завершение работы messaging service...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Сервер принудительно остановлен: %v", err)
		}
		log.Println("Messaging service остановлен")
	case err := <-errChan:
		log.Fatal("Ошибка сервера:", err)
	}
}
