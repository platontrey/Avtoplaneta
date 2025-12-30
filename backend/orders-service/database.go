package main

import (
	"log"

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

	// Автоматическая миграция схем заказов и запчастей
	if err := db.AutoMigrate(&Order{}, &OrderItem{}, &Part{}, &SalesHistory{}); err != nil {
		log.Fatal("Не удалось выполнить миграцию базы данных:", err)
	}
}