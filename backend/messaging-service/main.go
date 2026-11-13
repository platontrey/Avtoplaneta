package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/tracing"
)

func main() {
	// Установка количества OS-тредов для оптимизации под доступное количество ядер
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := godotenv.Load("../../.env"); err != nil {
		slog.Debug("No .env file found")
	}

	logFmt := logger.FormatJSON
	if os.Getenv("NODE_ENV") != "production" {
		logFmt = logger.FormatText
	}
	logger.Init(logger.Config{
		Level:       os.Getenv("LOG_LEVEL"),
		Format:      logFmt,
		ServiceName: "messaging-service",
	})

	// Инициализация OpenTelemetry Distributed Tracing
	tp, err := tracing.InitTracer("messaging-service")
	if err != nil {
		slog.Warn("Failed to init tracer for messaging-service", logger.Err(err))
	} else {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				slog.Error("Error shutting down tracer provider", logger.Err(err))
			}
		}()
	}

	// Initialize database
	InitDatabase()

	// Initialize Redis
	InitRedis()

	// Запуск gRPC-сервера в отдельной горутине
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9084"
	}
	go func() {
		if err := StartGRPCServer(grpcPort); err != nil {
			slog.Error("gRPC server failed", logger.Err(err))
			os.Exit(1)
		}
	}()

	// Создание контекста с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		slog.Info("Получен сигнал завершения, начинаем graceful shutdown...")
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
	r.Static("/api/messaging/uploads", "./uploads")
	r.Static("/api/v1/messaging/uploads", "./uploads")

	// Set API_BASE_URL
	if os.Getenv("MESSAGING_SERVICE_URL") != "" {
		os.Setenv("API_BASE_URL", os.Getenv("MESSAGING_SERVICE_URL"))
	} else {
		os.Setenv("API_BASE_URL", "http://localhost:8084")
	}

	// Start WebSocket Hub
	go globalHub.Run()

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
		slog.Info("Messaging service starting on port 8084")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			errChan <- err
		}
	}()

	// Ожидание сигнала отмены или ошибки
	select {
	case <-ctx.Done():
		slog.Info("Завершение работы messaging service...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("Сервер принудительно остановлен", logger.Err(err))
		}
		slog.Info("Messaging service остановлен")
	case err := <-errChan:
		slog.Error("Ошибка сервера", logger.Err(err))
		os.Exit(1)
	}
}
