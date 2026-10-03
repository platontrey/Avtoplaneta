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
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

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
	// Установка количества OS-тредов для оптимизации под доступное количество ядер
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Initialize OpenTelemetry Tracer
	tp, err := tracing.InitTracer("parts-service")
	if err != nil {
		logrus.WithError(err).Fatal("failed to initialize tracer")
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logrus.WithError(err).Error("failed to shutdown tracer")
		}
	}()

	// Установка режима Gin в Release для продакшена
	gin.SetMode(gin.ReleaseMode)

	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.LoadConfig()
	dbPool, err := repository.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbPool.Close()

	redisClient := initRedis(cfg)
	defer redisClient.Close()

	// Создание контекста с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация gRPC клиентов для межсервисной коммуникации
	grpc.InitGRPCClients()
	defer grpc.CloseGRPCClients()

	// Инициализация Elasticsearch
	var esClient search.ElasticsearchClient
	if err := search.InitElasticsearch(cfg.ElasticsearchURL); err != nil {
		log.Printf("Предупреждение: Не удалось инициализировать Elasticsearch: %v", err)
		log.Println("Продолжаем без функциональности Elasticsearch")
		esClient = nil
	} else {
		// Создание индекса запчастей, если он не существует
		if err := search.CreatePartsIndex(); err != nil {
			log.Printf("Предупреждение: Не удалось создать индекс запчастей: %v", err)
		}
		esClient = search.NewElasticsearchAdapter()
	}

	// Создание зависимостей с dependency injection
	repo := repository.NewPartRepository(dbPool)
	userDir := service.NewUserDirectory(grpc.GetAuthClient())
	svc := service.NewInventoryService(repo, esClient, cfg, userDir)
	svc.SetMonthlySalesProvider(grpc.GetMonthlySalesGRPC)

	if esClient != nil {
		if err := svc.ReindexAllParts(ctx); err != nil {
			log.Printf("Предупреждение: Не удалось переиндексировать запчасти: %v", err)
		}
	}

	partCatalog, err := catalog.LoadPartCatalog()
	if err != nil {
		log.Fatalf("Не удалось загрузить каталог шаблонов запчастей: %v", err)
	}
	vehicleCatalog, err := catalog.LoadVehicleCatalog()
	if err != nil {
		log.Fatalf("Не удалось загрузить справочник марок и моделей: %v", err)
	}
	defectReports := service.NewDefectReportWorkflow(
		partCatalog,
		svc,
	)
	h := handler.NewHandler(svc, partCatalog, vehicleCatalog, defectReports)

	// Запуск gRPC-сервера в отдельной горутине
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9081"
	}
	go func() {
		if err := grpc.StartGRPCServer(svc, defectReports, grpcPort); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Создание и запуск consumer событий
	eventConsumer := events.NewEventConsumer(redisClient, svc)
	go func() {
		if err := eventConsumer.Start(ctx); err != nil {
			log.Printf("Ошибка consumer событий: %v", err)
		}
	}()

	// Обработка сигналов для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Получен сигнал завершения, начинаем graceful shutdown...")
		cancel()
	}()

	r := gin.Default()
	r.Use(otelgin.Middleware("parts-service"))
	r.Use(handler.CORSMiddleware())
	handler.SetupRoutes(r, h)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	errChan := make(chan error, 1)
	go func() {
		log.Printf("Сервис запчастей запускается на порту %s", cfg.Port)

		certFile := "../../cert.pem"
		keyFile := "../../key.pem"
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			certFile = "cert.pem"
			keyFile = "key.pem"
		}

		if os.Getenv("NODE_ENV") == "production" {
			if _, err := os.Stat(certFile); err == nil {
				log.Println("Запуск с TLS...")
				if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
					errChan <- err
				}
			} else {
				log.Println("Сертификаты не найдены, запуск без TLS...")
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					errChan <- err
				}
			}
		} else {
			log.Println("Режим разработки: запуск без TLS...")
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errChan <- err
			}
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Завершение работы сервиса запчастей...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Сервер принудительно остановлен: %v", err)
		}
		log.Println("Сервис запчастей остановлен")
	case err := <-errChan:
		log.Fatal("Ошибка сервера:", err)
	}
}

func initRedis(cfg *config.Config) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURL,
		Password: cfg.RedisPassword,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Redis")
	}

	logrus.Info("Redis connected successfully")
	return redisClient
}
