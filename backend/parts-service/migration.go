package main

import (
	"log"

	"gorm.io/gorm"
)

// RunMigrations выполняет миграции базы данных
func RunMigrations(db *gorm.DB) {
	// Проверяем существование старого поля photo
	var hasOldPhotoColumn bool
	db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name='parts' AND column_name='photo' AND data_type='text'
		)
	`).Scan(&hasOldPhotoColumn)

	if hasOldPhotoColumn {
		log.Println("Миграция: Найдено старое поле photo, копируем данные в photos")
		// Копируем данные из старого поля photo в новое поле photos
		if err := db.Exec(`
			UPDATE parts
			SET photos = CASE
				WHEN photo IS NOT NULL AND photo != '' THEN jsonb_build_array(photo)
				ELSE '[]'::jsonb
			END
			WHERE photos IS NULL OR photos = '[]'::jsonb
		`).Error; err != nil {
			log.Printf("Ошибка копирования данных photo в photos: %v", err)
		} else {
			log.Println("Миграция: Данные photo скопированы в photos")
		}

		// Удаляем старое поле photo после успешного копирования
		if err := db.Exec(`ALTER TABLE parts DROP COLUMN IF EXISTS photo`).Error; err != nil {
			log.Printf("Ошибка удаления столбца photo: %v", err)
		} else {
			log.Println("Миграция: Старое поле photo удалено")
		}
	} else {
		log.Println("Миграция: Старое поле photo не найдено, миграция не требуется")
	}

	// Создаем индексы для оптимизации производительности
	createIndexes(db)

	log.Println("Миграции выполнены успешно")
}

// createIndexes создает необходимые индексы для оптимизации запросов
func createIndexes(db *gorm.DB) {
	log.Println("Миграция: Создание индексов для оптимизации производительности")

	// Включаем расширение pg_trgm для GIN индексов
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS pg_trgm`).Error; err != nil {
		log.Printf("Ошибка создания расширения pg_trgm: %v", err)
		return
	}

	indexes := []struct {
		name string
		sql  string
	}{
		// GIN индексы для полнотекстового поиска (ILIKE)
		{"idx_parts_name_gin", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_name_gin ON parts USING gin (name gin_trgm_ops)"},
		{"idx_parts_description_gin", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_description_gin ON parts USING gin (description gin_trgm_ops)"},

		// B-tree индексы для фильтров
		{"idx_parts_category", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_category ON parts (category)"},
		{"idx_parts_brand", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_brand ON parts (brand)"},
		{"idx_parts_model", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_model ON parts (model)"},
		{"idx_parts_location", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_location ON parts (location)"},
		{"idx_parts_salesman", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_salesman ON parts (salesman)"},
		{"idx_parts_status", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_status ON parts (status)"},
		{"idx_parts_supplier_code", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_supplier_code ON parts (supplier_code)"},
		{"idx_parts_to_delete_at", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_to_delete_at ON parts (to_delete_at)"},
		{"idx_parts_quantity", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_quantity ON parts (quantity)"},

		// Составной индекс для часто используемых фильтров
		{"idx_parts_status_quantity", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_status_quantity ON parts (status, quantity)"},
		{"idx_parts_category_brand", "CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_parts_category_brand ON parts (category, brand)"},
	}

	for _, idx := range indexes {
		if err := db.Exec(idx.sql).Error; err != nil {
			log.Printf("Ошибка создания индекса %s: %v", idx.name, err)
		} else {
			log.Printf("Индекс %s создан успешно", idx.name)
		}
	}

	log.Println("Миграция: Создание индексов завершено")
}