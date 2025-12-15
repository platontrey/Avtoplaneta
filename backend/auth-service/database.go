package main

import (
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// InitDB инициализирует подключение к базе данных и выполняет миграцию
func InitDB(config *Config) {
	var err error

	db, err = gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Millisecond * 200, // Логировать запросы медленнее 200ms
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	})
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	// Автоматическая миграция схемы пользователей и логов активности
	if err := db.AutoMigrate(&User{}, &UserActivityLog{}); err != nil {
		log.Fatal("Не удалось выполнить миграцию базы данных:", err)
	}
}

// CreateDefaultUser создает пользователя по умолчанию, если пользователей нет
func CreateDefaultUser() {
	// Проверить, есть ли уже пользователи
	var count int64
	db.Model(&User{}).Count(&count)
	if count > 0 {
		log.Println("Пользователи уже существуют, пропускаем создание default пользователя")
		return
	}

	// Создать default пользователя
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Не удалось хэшировать пароль для default пользователя: %v", err)
		return
	}

	defaultUser := User{
		Email:    "bibidatrbib@gmail.com",
		Name:     "Test User",
		Provider: "local",
		Role:     "admin",
		Password: string(hashedPassword),
	}

	if err := db.Create(&defaultUser).Error; err != nil {
		log.Printf("Не удалось создать default пользователя: %v", err)
		return
	}

	log.Println("Default пользователь создан: bibidatrbib@gmail.com / password123")
}