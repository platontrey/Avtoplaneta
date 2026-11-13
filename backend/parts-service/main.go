package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"avtoplaneta/pkg/httpserver"
	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/redisclient"
	"avtoplaneta/pkg/tracing"
	"parts-service/internal/catalog"
	"parts-service/internal/config"
	"parts-service/internal/events"
	"parts-service/internal/grpc"
	"parts-service/internal/handler"
	"parts-service/internal/outbox"
	"parts-service/internal/repository"
	"parts-service/internal/search"
	"parts-service/internal/service"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	gin.SetMode(gin.ReleaseMode)
	_ = godotenv.Load("../../.env")

	cfg := config.LoadConfig()

	// Инициализация структурированного логирования
	logger.Init(logger.Config{
		Level:       cfg.LogLevel,
		Format:      logger.Format(cfg.LogFormat),
		ServiceName: "parts-service",
	})

	tp, err := tracing.InitTracer("parts-service")
	if err != nil {
		slog.Error("failed to initialize tracer", logger.Err(err))
		os.Exit(1)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown tracer", logger.Err(err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := repository.InitDB(cfg)
	if err != nil {
		slog.Error("Failed to initialize database", logger.Err(err))
		os.Exit(1)
	}
	defer dbPool.Close()

	redisClient, err := redisclient.New(cfg.RedisURL, cfg.RedisPassword)
	if err != nil {
		slog.Error("Failed to connect to Redis", logger.Err(err))
		os.Exit(1)
	}
	defer redisClient.Close()

	grpc.InitGRPCClients()
	defer grpc.CloseGRPCClients()

	// Инициализация Elasticsearch
	var esClient search.ElasticsearchClient
	if err := search.InitElasticsearch(cfg.ElasticsearchURL); err != nil {
		slog.Warn("Elasticsearch is unavailable, running without full-text search", logger.Err(err))
	} else {
		if err := search.CreatePartsIndex(); err != nil {
			slog.Warn("Failed to create parts index", logger.Err(err))
		}
		esClient = search.NewElasticsearchAdapter()
	}

	// Сборка зависимостей (Dependency Injection)
	repo := repository.NewPartRepository(dbPool)
	userDir := service.NewUserDirectory(grpc.GetAuthClient())
	svc := service.NewInventoryService(repo, esClient, cfg, userDir)
	svc.SetMonthlySalesProvider(grpc.GetMonthlySalesGRPC)

	if esClient != nil {
		if err := svc.ReindexAllParts(ctx); err != nil {
			slog.Warn("Failed to reindex parts", logger.Err(err))
		}
	}

	partCatalog, err := catalog.LoadPartCatalog()
	if err != nil {
		slog.Error("Failed to load parts catalog", logger.Err(err))
		os.Exit(1)
	}
	vehicleCatalog, err := catalog.LoadVehicleCatalog()
	if err != nil {
		slog.Error("Failed to load vehicle catalog", logger.Err(err))
		os.Exit(1)
	}
	defectReports := service.NewDefectReportWorkflow(partCatalog, svc)
	h := handler.NewHandler(svc, partCatalog, vehicleCatalog, defectReports)

	// Регистрация readiness проверок (Postgres и Redis)
	h.HealthChecker().AddPinger("postgres", dbPool)
	if redisClient != nil {
		h.HealthChecker().AddReadinessCheck("redis", func(checkCtx context.Context) error {
			return redisClient.Ping(checkCtx).Err()
		})
	}

	// Запуск gRPC-сервера
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9081"
	}
	go func() {
		if err := grpc.StartGRPCServer(svc, defectReports, grpcPort); err != nil {
			slog.Error("gRPC server failed", logger.Err(err))
			os.Exit(1)
		}
	}()

	// Запуск потребителя событий Redis Streams
	eventConsumer := events.NewEventConsumer(redisClient, svc)
	go func() {
		if err := eventConsumer.Start(ctx); err != nil {
			slog.Error("Redis event consumer error", logger.Err(err))
		}
	}()

	// Запуск Transactional Outbox диспетчера с LISTEN/NOTIFY для суб-миллисекундной синхронизации ES и кэша
	outboxDispatcher := outbox.NewDispatcher(dbPool, repo, esClient, redisClient)
	go func() {
		if err := outboxDispatcher.Start(ctx); err != nil {
			slog.Error("Outbox dispatcher error", logger.Err(err))
		}
	}()

	// Настройка HTTP роутера
	r := gin.Default()
	r.Use(otelgin.Middleware("parts-service"))
	r.Use(handler.CORSMiddleware())
	handler.SetupRoutes(r, h)

	// Запуск HTTP-сервера с поддержкой TLS и graceful shutdown
	if err := httpserver.Run(ctx, cfg.Port, r); err != nil {
		slog.Error("HTTP server failed", logger.Err(err))
		os.Exit(1)
	}
}
