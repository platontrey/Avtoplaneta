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

	log.Println("Миграции выполнены успешно")
}