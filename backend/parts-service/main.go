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

	"avtoplaneta/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var redisClient *redis.Client

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

	config := LoadConfig()
	InitDB(config)
	InitRedis(config)

	// Создание контекста с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer closeGRPCClients()

	// Инициализация Elasticsearch
	var esClient ElasticsearchClient
	if err := InitElasticsearch(config.ElasticsearchURL); err != nil {
		log.Printf("Предупреждение: Не удалось инициализировать Elasticsearch: %v", err)
		log.Println("Продолжаем без функциональности Elasticsearch")
		esClient = nil
	} else {
		// Создание индекса запчастей, если он не существует
		if err := CreatePartsIndex(); err != nil {
			log.Printf("Предупреждение: Не удалось создать индекс запчастей: %v", err)
		} else {
			// Переиндексация всех существующих запчастей
			if err := ReindexAllParts(); err != nil {
				log.Printf("Предупреждение: Не удалось переиндексировать запчасти: %v", err)
			}
		}
		esClient = NewElasticsearchAdapter()
	}

	// Создание зависимостей с dependency injection
	repo := NewPartRepository(dbPool)
	service := NewInventoryService(repo, esClient, config, newUserDirectory())
	partCatalog, err := LoadPartCatalog()
	if err != nil {
		log.Fatalf("Не удалось загрузить каталог шаблонов запчастей: %v", err)
	}
	vehicleCatalog, err := LoadVehicleCatalog()
	if err != nil {
		log.Fatalf("Не удалось загрузить справочник марок и моделей: %v", err)
	}
	defectReports := NewDefectReportWorkflow(
		partCatalog,
		NewRedisDefectReportEventPublisher(redisClient),
	)
	handler := NewHandler(service, partCatalog, vehicleCatalog, defectReports)

	// Запуск gRPC-сервера в отдельной горутине
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9081"
	}
	go func() {
		if err := StartGRPCServer(service, defectReports, grpcPort); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Инициализация gRPC клиентов для межсервисной коммуникации
	initGRPCClients()

	// Создание и запуск consumer событий
	eventConsumer := NewEventConsumer(redisClient, service)
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

	// Запуск планировщика автоматической генерации XML с контекстом
	go StartXMLGenerationScheduler(ctx)

	r := gin.Default()
	r.Use(otelgin.Middleware("parts-service"))

	// CORS middleware для кросс-доменных запросов
	r.Use(CORSMiddleware())

	// Настройка маршрутов с handler
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
		log.Printf("Сервис запчастей запускается на порту %s", config.Port)

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

	// Ожидание сигнала отмены или ошибки
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

// InitRedis инициализирует подключение к Redis
func InitRedis(config *Config) {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     config.RedisURL,
		Password: config.RedisPassword,
	})

	// Проверяем подключение
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Redis")
	}

	logrus.Info("Redis connected successfully")
}
