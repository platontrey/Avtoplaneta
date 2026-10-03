package main

import (
	"context"
	"errors"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"avtoplaneta/pkg/httpserver"
	"avtoplaneta/pkg/redisclient"
	"orders-service/internal/config"
	"orders-service/internal/events"
	"orders-service/internal/grpc"
	"orders-service/internal/handler"
	"orders-service/internal/repository"
	"orders-service/internal/service"
)

func main() {
	// Оптимизация под количество ядер процессора
	runtime.GOMAXPROCS(runtime.NumCPU())

	gin.SetMode(gin.ReleaseMode)
	_ = godotenv.Load("../../.env")

	cfg := config.LoadConfig()

	// Контекст для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Инициализация базы данных и миграций
	dbPool, err := repository.InitDB(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize orders database")
	}
	defer dbPool.Close()

	// Инициализация клиента Redis
	redisClient, err := redisclient.New(cfg.RedisURL, cfg.RedisPassword)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	// gRPC клиент к parts-service
	partsClient, err := grpc.NewPartsGRPCClient(cfg.PartsServiceGRPCURL)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to connect to parts-service via gRPC at %s", cfg.PartsServiceGRPCURL)
	}
	defer partsClient.Close()

	// Сборка зависимостей (Dependency Injection)
	orderRepo := repository.NewOrderRepository(dbPool)
	partRepo := repository.NewPartRepositoryForOrders(partsClient)
	cacheService := repository.NewCacheService(redisClient)
	eventPublisher := events.NewEventPublisher(redisClient)
	ordersService := service.NewOrdersService(orderRepo, partRepo, cacheService, eventPublisher)
	h := handler.NewHandler(ordersService, eventPublisher)

	// Запуск потребителя событий Redis Streams в фоне
	eventConsumer := events.NewRedisEventConsumer(redisClient, orderRepo, cacheService)
	go func() {
		if err := eventConsumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logrus.WithError(err).Error("Redis event consumer failed in orders-service")
		}
	}()

	// Запуск gRPC-сервера в фоне
	go func() {
		if err := grpc.StartGRPCServer(ordersService, eventPublisher, cfg.GRPCPort); err != nil {
			logrus.WithError(err).Fatal("orders-service gRPC server failed")
		}
	}()

	// Настройка HTTP роутера Gin
	r := gin.Default()
	r.Use(handler.CORSMiddleware())
	handler.SetupRoutes(r, h)

	// Запуск HTTP-сервера с поддержкой TLS и graceful shutdown
	if err := httpserver.Run(ctx, cfg.Port, r); err != nil {
		logrus.WithError(err).Fatal("orders-service HTTP server failed")
	}
}
