package main

import (
	"context"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"avtoplaneta/pkg/httpserver"
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

	// 2. Загрузка конфигурации
	cfg := config.LoadConfig()

	// 3. Контекст graceful shutdown по сигналам ОС
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logrus.WithFields(logrus.Fields{
		"http_port": cfg.Port,
		"grpc_port": cfg.GRPCPort,
		"env":       cfg.NodeEnv,
	}).Info("Инициализация template-service...")

	// 4. Подключение к PostgreSQL и запуск встроенных миграций
	dbPool, err := repository.InitDB(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("Не удалось инициализировать базу данных PostgreSQL")
	}
	defer dbPool.Close()

	// 5. Сборка слоев (Dependency Injection)
	itemRepo := repository.NewItemRepository(dbPool)
	itemService := service.NewItemService(itemRepo)
	h := handler.NewHandler(itemService)

	// 6. Настройка HTTP-роутера Gin
	router := gin.New()
	handler.SetupRoutes(router, h)

	// 7. Конкурентный запуск HTTP и gRPC серверов через errgroup
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
		logrus.WithError(err).Error("Завершение работы сервиса с ошибкой")
	} else {
		logrus.Info("Сервис успешно остановлен (Graceful shutdown завершен)")
	}
}
