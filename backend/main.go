package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// @title Avtoplaneta Backend API
// @version 1.0
// @description API для системы управления автозапчастями Avtoplaneta
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.avtoplaneta.com/support
// @contact.email support@avtoplaneta.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @externalDocs.description OpenAPI
// @externalDocs.url https://swagger.io/resources/open-api/

func main() {
	// Загрузка переменных окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		logrus.WithError(err).Warn("Не удалось загрузить .env файл, используются системные переменные")
	}

	// Настройка режима Gin
	if os.Getenv("NODE_ENV") == "production" {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Создание и запуск API Gateway
	gateway := NewGateway()

	logrus.Info("API Gateway запущен на порту :8080")
	if err := gateway.Run(":8080"); err != nil {
		logrus.WithError(err).Fatal("Не удалось запустить API Gateway")
	}
}
