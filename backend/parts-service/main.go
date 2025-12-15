package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
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

	// Запуск планировщика автоматической генерации XML
	go StartXMLGenerationScheduler()

	r := gin.Default()

	// CORS middleware для кросс-доменных запросов
	r.Use(CORSMiddleware())

	// Настройка маршрутов с handler
	SetupRoutes(r, handler)

	log.Printf("Сервис запчастей запускается на порту %s", config.Port)
	if err := r.Run(":" + config.Port); err != nil {
		log.Fatal("Не удалось запустить сервер:", err)
	}
}
