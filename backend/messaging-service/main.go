package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"avtoplaneta/pkg/ginmw"
	"avtoplaneta/pkg/httpserver"
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

	// Контекст для graceful shutdown всего сервиса
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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
		if err := StartGRPCServer(ctx, grpcPort); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("gRPC server failed", logger.Err(err))
		}
	}()

	r := gin.Default()
	r.Use(ginmw.CORS())

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

	// Routes
	setupRoutes(r)

	// Запуск HTTP-сервера через общий пакет httpserver (graceful shutdown + TLS)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	if err := httpserver.Run(ctx, port, r); err != nil {
		slog.Error("Messaging service HTTP server failed", logger.Err(err))
		os.Exit(1)
	}
}
