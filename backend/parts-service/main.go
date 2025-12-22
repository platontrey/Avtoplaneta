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
)

func main() {
	// Установка количества OS-тредов
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Установка режима Gin в Release для продакшена
	gin.SetMode(gin.ReleaseMode)

	config := LoadConfig()
	InitDB(config)

	// Инициализация Elasticsearch
	var esClient ElasticsearchClient
	if err := InitElasticsearch(); err != nil {
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
	repo := NewPartRepository(db)
	service := NewInventoryService(repo, esClient)
	handler := NewHandler(service)

	// Создание контекста с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
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
