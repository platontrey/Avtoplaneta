package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"avtoplaneta/pkg/httpserver"
	"avtoplaneta/pkg/redisclient"
	"avtoplaneta/pkg/tracing"
	"parts-service/internal/catalog"
	"parts-service/internal/config"
	"parts-service/internal/events"
	"parts-service/internal/grpc"
	"parts-service/internal/handler"
	"parts-service/internal/repository"
	"parts-service/internal/search"
	"parts-service/internal/service"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	tp, err := tracing.InitTracer("parts-service")
	if err != nil {
		logrus.WithError(err).Fatal("failed to initialize tracer")
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logrus.WithError(err).Error("failed to shutdown tracer")
		}
	}()

	gin.SetMode(gin.ReleaseMode)
	_ = godotenv.Load("../../.env")

	cfg := config.LoadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbPool, err := repository.InitDB(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize database")
	}
	defer dbPool.Close()

	redisClient, err := redisclient.New(cfg.RedisURL, cfg.RedisPassword)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisClient.Close()

	grpc.InitGRPCClients()
	defer grpc.CloseGRPCClients()

	// Инициализация Elasticsearch
	var esClient search.ElasticsearchClient
	if err := search.InitElasticsearch(cfg.ElasticsearchURL); err != nil {
		logrus.WithError(err).Warn("Elasticsearch is unavailable, running without full-text search")
	} else {
		if err := search.CreatePartsIndex(); err != nil {
			logrus.WithError(err).Warn("Failed to create parts index")
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
			logrus.WithError(err).Warn("Failed to reindex parts")
		}
	}

	partCatalog, err := catalog.LoadPartCatalog()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load parts catalog")
	}
	vehicleCatalog, err := catalog.LoadVehicleCatalog()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load vehicle catalog")
	}
	defectReports := service.NewDefectReportWorkflow(partCatalog, svc)
	h := handler.NewHandler(svc, partCatalog, vehicleCatalog, defectReports)

	// Запуск gRPC-сервера
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9081"
	}
	go func() {
		if err := grpc.StartGRPCServer(svc, defectReports, grpcPort); err != nil {
			logrus.WithError(err).Fatal("gRPC server failed")
		}
	}()

	// Запуск потребителя событий Redis Streams
	eventConsumer := events.NewEventConsumer(redisClient, svc)
	go func() {
		if err := eventConsumer.Start(ctx); err != nil {
			logrus.WithError(err).Error("Redis event consumer error")
		}
	}()

	// Настройка HTTP роутера
	r := gin.Default()
	r.Use(otelgin.Middleware("parts-service"))
	r.Use(handler.CORSMiddleware())
	handler.SetupRoutes(r, h)

	// Запуск HTTP-сервера с поддержкой TLS и graceful shutdown
	if err := httpserver.Run(ctx, cfg.Port, r); err != nil {
		logrus.WithError(err).Fatal("HTTP server failed")
	}
}
