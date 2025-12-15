package main

import "os"

// Config содержит конфигурационные параметры приложения
type Config struct {
	DatabaseURL string
	Port        string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	config := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
	}

	// Значения по умолчанию
	if config.DatabaseURL == "" {
		config.DatabaseURL = "host=localhost user=postgres password=qewret123 dbname=autoplanet port=5432 sslmode=disable"
	}
	if config.Port == "" {
		config.Port = "8081"
	}

	return config
}