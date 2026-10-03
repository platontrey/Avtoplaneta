package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	partsv1 "avtoplaneta/gen/parts/v1"
	"avtoplaneta/pkg/httpserver"
	"avtoplaneta/pkg/redisclient"
	"avtoplaneta/pkg/tracing"

	"auth-service/internal/config"
	"auth-service/internal/events"
	grpcserver "auth-service/internal/grpc"
	"auth-service/internal/handler"
	"auth-service/internal/repository"
	"auth-service/internal/security"
	"auth-service/internal/service"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Инициализация трассировки OpenTelemetry
	tp, err := tracing.InitTracer("auth-service")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize OpenTelemetry tracer")
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logrus.WithError(err).Error("Failed to shutdown tracer")
		}
	}()

	gin.SetMode(gin.ReleaseMode)

	// Контекст для graceful shutdown всего сервиса
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Загрузка конфигурации
	cfg := config.LoadConfig()

	// Подключение к базе данных PostgreSQL
	dbPool, err := repository.InitDB(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize database")
	}
	defer dbPool.Close()

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(dbPool)
	activityRepo := repository.NewActivityLogRepository(dbPool)

	// Создание дефолтного администратора при первом запуске
	repository.CreateDefaultUser(ctx, userRepo)

	// Сессии и безопасность
	sessionStore, err := security.NewCookieSessionStore(cfg.SessionSecret)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize cookie session store")
	}

	csrfManager := security.NewCSRFManager()
	csrfManager.StartCleanup(ctx)
	security.InitGoogleOAuth("")
	rateLimiter := security.NewMemoryRateLimiter()

	// Подключение к Redis для публикации событий пользователей
	var eventPublisher events.UserEventPublisher
	if cfg.RedisURL != "" {
		rdb, err := redisclient.New(cfg.RedisURL, "")
		if err != nil {
			logrus.WithError(err).Warn("Failed to connect to Redis for user events")
		} else {
			eventPublisher = events.NewRedisUserEventPublisher(rdb)
			defer rdb.Close()
		}
	}

	// Сервисный слой
	authService := service.NewAuthService(
		userRepo,
		activityRepo,
		sessionStore,
		rateLimiter,
		csrfManager,
		eventPublisher,
	)

	// HTTP Handler с явным DI
	h := handler.NewHandler(authService, sessionStore, csrfManager, userRepo, cfg)

	// Подключение к parts-service по gRPC
	if cfg.PartsGRPCAddr != "" {
		partsConn, err := googlegrpc.NewClient(cfg.PartsGRPCAddr, googlegrpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logrus.WithError(err).Warn("Failed to create gRPC connection to parts-service")
		} else {
			h.SetPartsClient(partsv1.NewPartsServiceClient(partsConn))
			logrus.WithField("addr", cfg.PartsGRPCAddr).Info("Connected to parts-service via gRPC")
			defer partsConn.Close()
		}
	}

	// Запуск gRPC сервера
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9083"
	}
	grpcSrv, err := grpcserver.StartGRPCServer(authService, sessionStore, cfg, grpcPort)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to start gRPC server on port %s", grpcPort)
	}
	defer grpcSrv.GracefulStop()

	// Настройка HTTP сервера
	r := gin.Default()
	r.Use(otelgin.Middleware("auth-service"))
	_ = r.SetTrustedProxies([]string{"127.0.0.1"})
	r.Use(h.CSRFMiddleware)

	handler.SetupRoutes(r, h)

	// Запуск HTTP сервера через общий пакет httpserver (graceful shutdown + TLS)
	if err := httpserver.Run(ctx, cfg.Port, r); err != nil {
		logrus.WithError(err).Fatal("Server error")
	}
}
