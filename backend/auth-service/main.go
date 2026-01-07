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
)

func main() {
	// Установка количества OS-тредов для оптимизации под доступное количество ядер
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Установка режима Gin в Release для продакшена
	gin.SetMode(gin.ReleaseMode)

	config := LoadConfig()
	InitDB(config)
	CreateDefaultUser()

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

	InitAuth(ctx, config)

	// Создаем зависимости
	userRepo := NewUserRepository(db)
	activityRepo := NewActivityLogRepository(db)
	authService := NewAuthService(userRepo, activityRepo, store, nil) // Rate limiter пока не реализован
	handler := NewHandler(authService, config)

	log.Println("Сервис аутентификации готов к работе с пользователями.")

	r := gin.Default()

	// Установить доверенные прокси для безопасности
	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Printf("Ошибка установки доверенных прокси: %v", err)
	}

	// Setup middleware
	r.Use(CORSMiddleware(config))
	r.Use(csrfMiddleware)

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
		log.Printf("Сервис аутентификации запускается на порту %s", config.Port)

		certFile := "../../cert.pem"
		keyFile := "../../key.pem"
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			certFile = "cert.pem"
			keyFile = "key.pem"
		}

		if os.Getenv("NODE_ENV") == "production" {
			if _, err := os.Stat(certFile); err == nil {
				log.Println("Запуск с TLS...")
				if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && !errors.Is(http.ErrServerClosed, err) {
					errChan <- err
				}
			} else {
				log.Println("Сертификаты не найдены, запуск без TLS...")
				if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
					errChan <- err
				}
			}
		} else {
			log.Println("Режим разработки: запуск без TLS...")
			if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
				errChan <- err
			}
		}
	}()

	// Ожидание сигнала отмены или ошибки
	select {
	case <-ctx.Done():
		log.Println("Завершение работы сервиса аутентификации...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Сервер принудительно остановлен: %v", err)
		}
		log.Println("Сервис аутентификации остановлен")
	case err := <-errChan:
		log.Fatal("Ошибка сервера:", err)
	}
}
