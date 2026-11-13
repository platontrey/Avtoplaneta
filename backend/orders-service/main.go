package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"avtoplaneta/pkg/httpserver"
	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/redisclient"
	"avtoplaneta/pkg/tracing"
	"orders-service/internal/config"
	"orders-service/internal/events"
	"orders-service/internal/grpc"
	"orders-service/internal/handler"
	"orders-service/internal/outbox"
	"orders-service/internal/repository"
	"orders-service/internal/service"
)

func main() {
	// Оптимизация под количество ядер процессора
	runtime.GOMAXPROCS(runtime.NumCPU())

	gin.SetMode(gin.ReleaseMode)
	_ = godotenv.Load("../../.env")

	cfg := config.LoadConfig()

	// Инициализация структурированного логирования
	logger.Init(logger.Config{
		Level:  cfg.LogLevel,
		Format:      logger.Format(cfg.LogFormat),
		ServiceName: "orders-service",
	})

	// Инициализация OpenTelemetry Distributed Tracing
	tp, err := tracing.InitTracer("orders-service")
	if err != nil {
		slog.Warn("Failed to init tracer for orders-service", logger.Err(err))
	} else {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				slog.Warn("Error shutting down tracer provider", logger.Err(err))
			}
		}()
	}

	// Контекст для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Инициализация базы данных и миграций
	dbPool, err := repository.InitDB(cfg)
	if err != nil {
		slog.Error("Failed to initialize orders database", logger.Err(err))
		os.Exit(1)
	}
	defer dbPool.Close()

	// Инициализация клиента Redis
	redisClient, err := redisclient.New(cfg.RedisURL, cfg.RedisPassword)
	if err != nil {
		slog.Error("Failed to connect to Redis", logger.Err(err))
		os.Exit(1)
	}
	defer redisClient.Close()

	// gRPC клиент к parts-service
	partsClient, err := grpc.NewPartsGRPCClient(cfg.PartsServiceGRPCURL)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to connect to parts-service via gRPC at %s", cfg.PartsServiceGRPCURL), logger.Err(err))
		os.Exit(1)
	}
	defer partsClient.Close()

	// Сборка зависимостей (Dependency Injection)
	orderRepo := repository.NewOrderRepository(dbPool)
	partRepo := repository.NewPartRepositoryForOrders(partsClient)
	cacheService := repository.NewCacheService(redisClient)
	eventPublisher := events.NewEventPublisher(redisClient)
	ordersService := service.NewOrdersService(orderRepo, partRepo, cacheService, eventPublisher)
	h := handler.NewHandler(ordersService, eventPublisher)

	// Регистрация readiness проверок (Postgres и Redis)
	h.HealthChecker().AddPinger("postgres", dbPool)
	if redisClient != nil {
		h.HealthChecker().AddReadinessCheck("redis", func(checkCtx context.Context) error {
			return redisClient.Ping(checkCtx).Err()
		})
	}

	// Запуск потребителя событий Redis Streams в фоне
	eventConsumer := events.NewRedisEventConsumer(redisClient, orderRepo, cacheService)
	go func() {
		if err := eventConsumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("Redis event consumer failed in orders-service", logger.Err(err))
		}
	}()

	// Запуск Transactional Outbox диспетчера с LISTEN/NOTIFY
	outboxDispatcher := outbox.NewDispatcher(dbPool, orderRepo, partRepo, redisClient, cacheService)
	go func() {
		if err := outboxDispatcher.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("Orders outbox dispatcher failed", logger.Err(err))
		}
	}()

	// Запуск gRPC-сервера в фоне
	go func() {
		if err := grpc.StartGRPCServer(ordersService, eventPublisher, cfg.GRPCPort); err != nil {
			slog.Error("orders-service gRPC server failed", logger.Err(err))
			os.Exit(1)
		}
	}()

	// Настройка HTTP роутера Gin
	r := gin.Default()
	r.Use(handler.CORSMiddleware())
	handler.SetupRoutes(r, h)

	// Запуск HTTP-сервера с поддержкой TLS и graceful shutdown
	if err := httpserver.Run(ctx, cfg.Port, r); err != nil {
		slog.Error("orders-service HTTP server failed", logger.Err(err))
		os.Exit(1)
	}
}
