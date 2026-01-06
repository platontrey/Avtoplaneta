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

	// Выполнить дополнительные миграции (индексы)
	RunMigrations(db)
}

// RunMigrations выполняет дополнительные миграции для оптимизации
func RunMigrations(db *gorm.DB) {
	log.Println("Выполнение миграций для orders-service")

	indexes := []struct {
		name string
		sql  string
	}{
		// Индексы для order_items
		{"idx_order_items_order_id", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_order_items_order_id ON order_items (order_id)"},
		{"idx_order_items_part_id", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_order_items_part_id ON order_items (part_id)"},

		// Индексы для orders
		{"idx_orders_status", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_status ON orders (status)"},
		{"idx_orders_created_at", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_created_at ON orders (created_at DESC)"},
		{"idx_orders_auto_deleted", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_auto_deleted ON orders (auto_deleted)"},
		{"idx_orders_user_id", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_orders_user_id ON orders (user_id)"},

		// Индексы для parts (для orders-service)
		{"idx_parts_id_orders", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_id_orders ON parts (id)"},
		{"idx_parts_quantity_orders", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_quantity_orders ON parts (quantity)"},

		// Индексы для sales_history
		{"idx_sales_history_created_at", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sales_history_created_at ON sales_history (created_at DESC)"},
	}

	for _, idx := range indexes {
		if err := db.Exec(idx.sql).Error; err != nil {
			log.Printf("Ошибка создания индекса %s: %v", idx.name, err)
		} else {
			log.Printf("Индекс %s создан успешно", idx.name)
		}
	}

	log.Println("Миграции orders-service выполнены")
}