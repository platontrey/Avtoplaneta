package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	"avtoplaneta/pkg/httpserver"
	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/redisclient"
	"avtoplaneta/pkg/tracing"
	"template-service/internal/config"
	"template-service/internal/grpc"
	"template-service/internal/handler"
	"template-service/internal/repository"
	"template-service/internal/service"
)

func main() {
	// 1. Оптимизация под доступные ядра CPU
	runtime.GOMAXPROCS(runtime.NumCPU())
	gin.SetMode(gin.ReleaseMode)

	// 2. Загрузка конфигурации и логгера
	cfg := config.LoadConfig()

	// 3. Инициализация распределенной трассировки OpenTelemetry
	tp, err := tracing.InitTracer("template-service")
	if err != nil {
		slog.Warn("Трассировщик OpenTelemetry не запущен (продолжение без трассировки)", logger.Err(err))
	} else {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				slog.Error("Ошибка при остановке трассировщика", logger.Err(err))
			}
		}()
	}

	// 4. Контекст graceful shutdown по сигналам ОС
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("Инициализация template-service...",
		slog.String("http_port", cfg.Port),
		slog.String("grpc_port", cfg.GRPCPort),
		slog.String("env", cfg.NodeEnv),
	)

	// 5. Подключение к PostgreSQL и запуск встроенных миграций
	dbPool, err := repository.InitDB(cfg)
	if err != nil {
		slog.Error("Не удалось инициализировать базу данных PostgreSQL", logger.Err(err))
		os.Exit(1)
	}
	defer dbPool.Close()

	// 6. Подключение к Redis (кэш, события, распределенные блокировки)
	var redisClient *redis.Client
	if cfg.RedisURL != "" {
		client, err := redisclient.New(cfg.RedisURL, cfg.RedisPassword)
		if err != nil {
			slog.Warn("Не удалось подключиться к Redis (кэширование и события отключены)", logger.Err(err))
		} else {
			redisClient = client
			defer redisClient.Close()
			slog.Info("Подключение к Redis успешно установлено")
		}
	}

	// 7. Сборка слоев (Dependency Injection)
	itemRepo := repository.NewItemRepository(dbPool)
	itemService := service.NewItemService(itemRepo)
	h := handler.NewHandler(itemService)

	// Регистрация readiness проверок (Postgres и Redis)
	h.HealthChecker().AddPinger("postgres", dbPool)
	if redisClient != nil {
		h.HealthChecker().AddReadinessCheck("redis", func(checkCtx context.Context) error {
			return redisClient.Ping(checkCtx).Err()
		})
	}

	// 8. Настройка HTTP-роутера Gin
	router := gin.New()
	handler.SetupRoutes(router, h)

	// 9. Конкурентный запуск HTTP и gRPC серверов через errgroup
	g, gCtx := errgroup.WithContext(ctx)

	// Запуск gRPC-сервера
	grpcServer := grpc.NewServer(itemService, cfg.GRPCPort)
	g.Go(func() error {
		return grpcServer.Start(gCtx)
	})

	// Запуск HTTP-сервера с поддержкой TLS и graceful shutdown
	g.Go(func() error {
		return httpserver.Run(gCtx, cfg.Port, router)
	})

	if err := g.Wait(); err != nil {
		slog.Error("Завершение работы сервиса с ошибкой", logger.Err(err))
	} else {
		slog.Info("Сервис успешно остановлен (Graceful shutdown завершен)")
	}
}
