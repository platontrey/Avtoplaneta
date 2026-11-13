package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/httpserver"
	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/redisclient"
	"export-service/internal/config"
	"export-service/internal/events"
	"export-service/internal/grpc"
	"export-service/internal/handler"
	"export-service/internal/repository"
	"export-service/internal/service"
)

// export-service собирает и валидирует XML прайс-листы для классифайдов (Drom)
// с использованием инкрементального слепка каталога в bbolt.
func main() {
	cfg := config.LoadConfig()

	logFmt := logger.FormatJSON
	if os.Getenv("NODE_ENV") != "production" {
		logFmt = logger.FormatText
	}
	logger.Init(logger.Config{
		Level:       os.Getenv("LOG_LEVEL"),
		Format:      logFmt,
		ServiceName: "export-service",
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. gRPC клиенты к parts-service и auth-service
	clients, err := grpc.NewClients(cfg.PartsServiceGRPC, cfg.AuthServiceGRPC)
	if err != nil {
		slog.Error("Не удалось подключиться к зависимым сервисам", logger.Err(err))
		os.Exit(1)
	}
	defer clients.Close()

	// 2. Redis клиент через avtoplaneta/pkg/redisclient (с мягким fallback при отсутствии)
	var rdb *redis.Client
	if cfg.RedisURL != "" {
		client, err := redisclient.New(cfg.RedisURL, cfg.RedisPassword)
		if err != nil {
			slog.Warn("Не удалось подключиться к Redis, export-service работает на резервной локальной очереди", logger.Err(err))
		} else {
			rdb = client
			defer rdb.Close()
		}
	}

	// 3. Репозиторий инкрементального слепка bbolt
	snapshotRepo, err := repository.NewBoltSnapshotRepository(cfg.CatalogDBPath)
	if err != nil {
		slog.Error("Не удалось инициализировать базу слепка каталога bbolt", logger.Err(err))
		os.Exit(1)
	}
	defer snapshotRepo.Close()

	// 4. Доменный сервис сборки каталога и прайс-листов
	priceListSvc := service.NewPriceListService(clients, snapshotRepo, cfg)

	// 5. Очередь асинхронных задач выгрузки
	exportQueue := events.NewExportQueue(rdb, priceListSvc, cfg)
	exportQueue.StartWorker(ctx)

	// 6. Фоновые процессы: планировщик и слушатель изменений склада
	readyChan := make(chan struct{})
	if snapshotRepo.GetVersion() != "" {
		close(readyChan)
	}

	go service.StartPriceListScheduler(ctx, priceListSvc, readyChan)
	go events.StartCatalogEventListener(ctx, rdb, priceListSvc, readyChan)

	// 7. HTTP-обработчик и роуты
	h := handler.NewHandler(priceListSvc, exportQueue, cfg)
	if rdb != nil {
		h.HealthChecker().AddReadinessCheck("redis", func(checkCtx context.Context) error {
			return rdb.Ping(checkCtx).Err()
		})
	}

	router := gin.Default()
	handler.SetupRoutes(router, h)

	// 8. Запуск сервера с поддержкой graceful shutdown и TLS через avtoplaneta/pkg/httpserver
	slog.Info("Запуск export-service", slog.String("port", cfg.Port))
	if err := httpserver.Run(ctx, cfg.Port, router); err != nil {
		slog.Error("Ошибка работы HTTP-сервера", logger.Err(err))
		os.Exit(1)
	}
}
