package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func initDB() {
	var err error

	// Строка подключения к PostgreSQL
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Локальное подключение к PostgreSQL по умолчанию
		dsn = "host=localhost user=postgres password=qewret123 dbname=autoplanet port=5432 sslmode=disable"
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	// Автоматическая миграция схемы запчастей
	if err := db.AutoMigrate(&Part{}); err != nil {
		log.Fatal("Не удалось выполнить миграцию:", err)
	}
}

// reindexAllParts переиндексирует все существующие запчасти в Elasticsearch
func reindexAllParts() error {
	var parts []Part
	if err := db.Find(&parts).Error; err != nil {
		return fmt.Errorf("не удалось получить запчасти: %v", err)
	}

	fmt.Printf("Переиндексация %d запчастей...\n", len(parts))
	for _, part := range parts {
		if err := IndexPart(&part); err != nil {
			fmt.Printf("Предупреждение: Не удалось проиндексировать запчасть %d: %v\n", part.ID, err)
		} else {
			fmt.Printf("Успешно проиндексирована запчасть %d\n", part.ID)
		}
	}
	fmt.Println("Переиндексация завершена")
	return nil
}

// startXMLGenerationScheduler запускает планировщик автоматической генерации XML прайс-листа каждые 14 дней
func startXMLGenerationScheduler() {
	ticker := time.NewTicker(14 * 24 * time.Hour) // 14 дней
	defer ticker.Stop()

	// Генерируем XML сразу при запуске
	fmt.Println("Запуск начальной генерации XML прайс-листа...")
	generateXMLPriceList()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Автоматическая генерация XML прайс-листа каждые 14 дней...")
			generateXMLPriceList()
		}
	}
}

// generateXMLPriceList генерирует и сохраняет XML прайс-лист
func generateXMLPriceList() {
	parts, err := GetPartsForXML()
	if err != nil {
		fmt.Printf("Ошибка получения частей для автоматической генерации XML: %v\n", err)
		return
	}

	xmlData, err := GenerateXMLPriceList(parts)
	if err != nil {
		fmt.Printf("Ошибка генерации XML для автоматической генерации: %v\n", err)
		return
	}

	// Сохранить XML файл на сервере
	filename := "pricelist.xml"
	filepath := "./uploads/" + filename

	// Создать директорию uploads, если она не существует
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		fmt.Printf("Ошибка создания директории uploads для автоматической генерации: %v\n", err)
		return
	}

	// Записать файл
	if err := os.WriteFile(filepath, xmlData, 0644); err != nil {
		fmt.Printf("Ошибка сохранения XML файла для автоматической генерации: %v\n", err)
		return
	}

	fmt.Printf("Успешно автоматически сгенерирован и сохранен XML прайс-лист с %d предложениями\n", len(parts))
}

func main() {
	initDB()

	// Инициализация Elasticsearch
	if err := InitElasticsearch(); err != nil {
		log.Printf("Предупреждение: Не удалось инициализировать Elasticsearch: %v", err)
		log.Println("Продолжаем без функциональности Elasticsearch")
	} else {
		// Создание индекса запчастей, если он не существует
		if err := CreatePartsIndex(); err != nil {
			log.Printf("Предупреждение: Не удалось создать индекс запчастей: %v", err)
		} else {
			// Переиндексация всех существующих запчастей
			if err := reindexAllParts(); err != nil {
				log.Printf("Предупреждение: Не удалось переиндексировать запчасти: %v", err)
			}
		}
	}

	// Запуск планировщика автоматической генерации XML
	go startXMLGenerationScheduler()

	r := gin.Default()

	// CORS middleware для кросс-доменных запросов
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "false")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Основные маршруты сервиса запчастей
	r.GET("/api/inventory", getInventory)
	r.POST("/api/addpart", addPart)
	r.DELETE("/api/deletepart/:id", deletePart)
	r.PUT("/api/updatepart/:id", updatePart)
	r.POST("/api/uploadpartphoto/:id", uploadPartPhoto)
	r.DELETE("/api/deletepartphoto/:id", deletePartPhoto)
	r.POST("/api/markpartfordeletion/:id", markPartForDeletion)
	r.GET("/api/statistics", getStatistics)
	r.GET("/api/export/xml", exportXMLPriceList)
	r.POST("/api/export/drom", sendPriceListToDrom)

	// Маршруты для характеристик запчастей больше не нужны - характеристики хранятся в основной таблице Part

	// Маршруты для дефектных ведомостей
	r.POST("/api/defect-reports", createDefectReport)

	// Админ маршруты
	r.DELETE("/api/admin/delete-zero-quantity-parts/:supplier_code", deleteZeroQuantityPartsBySupplier)
	r.GET("/api/admin/supplier-codes", getSupplierCodes)
	r.DELETE("/api/admin/bulk-delete-parts", bulkDeleteParts)
	r.PUT("/api/admin/bulk-update-parts", bulkUpdateParts)

	// Статическое обслуживание файлов для загрузок
	r.Static("/uploads", "./uploads")

	log.Println("Сервис запчастей запускается на порту :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatal("Не удалось запустить сервер:", err)
	}
}

// Функция createDefaultTemplates больше не нужна - характеристики хранятся в основной таблице Part
