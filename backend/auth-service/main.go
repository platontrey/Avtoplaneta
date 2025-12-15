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
	CreateDefaultUser()
	InitAuth(config)

	// Создаем зависимости
	userRepo := NewUserRepository(db)
	activityRepo := NewActivityLogRepository(db)
	authService := NewAuthService(userRepo, activityRepo, store, nil) // Rate limiter пока не реализован
	handler := NewHandler(authService)

	log.Println("Сервис аутентификации готов к работе с пользователями.")

	r := gin.Default()

	// Установить доверенные прокси для безопасности
	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		log.Printf("Ошибка установки доверенных прокси: %v", err)
	}

	// Setup middleware
	r.Use(CORSMiddleware())
	r.Use(csrfMiddleware)

	// Setup routes с dependency injection
	SetupRoutes(r, handler)

	log.Printf("Сервис аутентификации запускается на порту %s", config.Port)
	if err := r.Run(":" + config.Port); err != nil {
		log.Fatal("Не удалось запустить сервис аутентификации:", err)
	}
}
