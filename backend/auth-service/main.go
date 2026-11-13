package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	partsv1 "avtoplaneta/gen/parts/v1"
	"avtoplaneta/pkg/grpcclient"
	"avtoplaneta/pkg/httpserver"
	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/redisclient"
	"avtoplaneta/pkg/tracing"

	"github.com/redis/go-redis/v9"

	"auth-service/internal/config"
	"auth-service/internal/events"
	grpcserver "auth-service/internal/grpc"
	"auth-service/internal/handler"
	"auth-service/internal/outbox"
	"auth-service/internal/repository"
	"auth-service/internal/security"
	"auth-service/internal/service"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	logFmt := logger.FormatJSON
	if os.Getenv("NODE_ENV") != "production" {
		logFmt = logger.FormatText
	}
	logger.Init(logger.Config{
		Level:       os.Getenv("LOG_LEVEL"),
		Format:      logFmt,
		ServiceName: "auth-service",
	})

	// Инициализация трассировки OpenTelemetry
	tp, err := tracing.InitTracer("auth-service")
	if err != nil {
		slog.Error("Failed to initialize OpenTelemetry tracer", logger.Err(err))
		os.Exit(1)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("Failed to shutdown tracer", logger.Err(err))
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
		slog.Error("Failed to initialize database", logger.Err(err))
		os.Exit(1)
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
		slog.Error("Failed to initialize cookie session store", logger.Err(err))
		os.Exit(1)
	}

	csrfManager := security.NewCSRFManager()
	csrfManager.StartCleanup(ctx)
	security.InitGoogleOAuth("")
	rateLimiter := security.NewMemoryRateLimiter()

	// Подключение к Redis для публикации событий пользователей
	var redisClient *redis.Client
	var eventPublisher events.UserEventPublisher
	if cfg.RedisURL != "" {
		rdb, err := redisclient.New(cfg.RedisURL, "")
		if err != nil {
			slog.Warn("Failed to connect to Redis for user events", logger.Err(err))
		} else {
			redisClient = rdb
			eventPublisher = events.NewRedisUserEventPublisher(rdb)
			defer rdb.Close()
		}
	}

	// Запуск Transactional Outbox диспетчера с LISTEN/NOTIFY
	outboxDispatcher := outbox.NewDispatcher(dbPool, userRepo, redisClient)
	go func() {
		if err := outboxDispatcher.Start(ctx); err != nil {
			slog.Error("Auth outbox dispatcher failed", logger.Err(err))
		}
	}()

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

	// Регистрация readiness проверок (Postgres и Redis)
	h.HealthChecker().AddPinger("postgres", dbPool)
	if redisClient != nil {
		h.HealthChecker().AddReadinessCheck("redis", func(checkCtx context.Context) error {
			return redisClient.Ping(checkCtx).Err()
		})
	}

	// Подключение к parts-service по gRPC
	if cfg.PartsGRPCAddr != "" {
		partsConn, err := grpcclient.NewClient(cfg.PartsGRPCAddr)
		if err != nil {
			slog.Warn("Failed to create gRPC connection to parts-service", logger.Err(err))
		} else {
			h.SetPartsClient(partsv1.NewPartsServiceClient(partsConn))
			slog.Info("Connected to parts-service via gRPC", slog.String("addr", cfg.PartsGRPCAddr))
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
		slog.Error("Failed to start gRPC server", logger.Err(err), slog.String("port", grpcPort))
		os.Exit(1)
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
		slog.Error("Server error", logger.Err(err))
		os.Exit(1)
	}
}
