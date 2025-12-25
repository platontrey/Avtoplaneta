package main

import "os"

// Config содержит конфигурационные параметры приложения
type Config struct {
	DatabaseURL string
	Port        string
	RedisURL    string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	config := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
		RedisURL:    os.Getenv("REDIS_URL"),
	}

	// Значения по умолчанию
	if config.DatabaseURL == "" {
		config.DatabaseURL = "host=localhost user=postgres password=qewret123 dbname=autoplanet port=5432 sslmode=disable"
	}
	if config.Port == "" {
		config.Port = "8081"
	}
	if config.RedisURL == "" {
		config.RedisURL = "127.0.0.1:6379"
	}

	return config
}