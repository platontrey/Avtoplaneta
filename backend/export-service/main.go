package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"avtoplaneta/pkg/redisclient"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// export-service собирает прайс-лист для Drom.
//
// Своей базы у сервиса нет: запчасти он читает из parts-service, реквизиты
// продавцов — из auth-service. Раньше это всё жило внутри parts-service, и тот
// ради выгрузки ходил в auth-service за ИНН — за персональными данными, которые
// складу не нужны ни для чего. Теперь ИНН знает только тот, кто его печатает.
func main() {
	config := LoadConfig()

	clients, err := NewClients(config.PartsServiceGRPC, config.AuthServiceGRPC)
	if err != nil {
		logrus.WithError(err).Fatal("Не удалось подключиться к сервисам")
	}
	defer clients.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var rdb *redis.Client
	if config.RedisURL != "" {
		client, err := redisclient.New(config.RedisURL, config.RedisPassword)
		if err != nil {
			logrus.WithError(err).Warn("Не удалось подключиться к Redis, export-service работает на резервной локальной очереди")
		} else {
			rdb = client
			defer rdb.Close()
		}
	}

	builder := NewPriceListBuilder(clients, config)
	queue := NewExportQueue(rdb, builder, config)
	queue.StartWorker(ctx)

	handler := NewHandler(builder, queue, config)

	router := gin.Default()
	SetupRoutes(router, handler)

	go StartPriceListScheduler(ctx, builder)

	server := &http.Server{
		Addr:              ":" + config.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logrus.WithField("port", config.Port).Info("export-service запущен")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.WithError(err).Fatal("Ошибка HTTP-сервера")
		}
	}()

	<-ctx.Done()
	logrus.Info("Остановка export-service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Error("Ошибка остановки HTTP-сервера")
	}
}
