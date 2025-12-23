package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// InitDB инициализирует подключение к базе данных и выполняет миграцию
func InitDB(config *Config) {
	var err error

	db, err = gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	// Настройка connection pooling для оптимизации производительности
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Не удалось получить SQL DB:", err)
	}

	// Настройки пула соединений
	sqlDB.SetMaxIdleConns(10)                 // Максимальное количество idle соединений
	sqlDB.SetMaxOpenConns(100)                // Максимальное количество открытых соединений
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // Максимальное время жизни соединения

	log.Println("Connection pooling настроен: MaxIdleConns=10, MaxOpenConns=100, ConnMaxLifetime=30m")

	// Автоматическая миграция схемы запчастей и earnings
	if err := db.AutoMigrate(&Part{}, &Earnings{}); err != nil {
		log.Fatal("Не удалось выполнить миграцию:", err)
	}

	// Выполнение кастомных миграций
	RunMigrations(db)
}

// ReindexAllParts переиндексирует все существующие запчасти в Elasticsearch
func ReindexAllParts() error {
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