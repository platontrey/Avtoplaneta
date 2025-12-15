package main

import "os"

// Config содержит конфигурационные параметры приложения
type Config struct {
	DatabaseURL   string
	SessionSecret string
	NodeEnv       string
	Port          string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	config := &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
		NodeEnv:       os.Getenv("NODE_ENV"),
		Port:          os.Getenv("PORT"),
	}

	// Значения по умолчанию
	if config.DatabaseURL == "" {
		config.DatabaseURL = "host=localhost user=postgres password=qewret123 dbname=autoplanet port=5432 sslmode=disable"
	}
	if config.SessionSecret == "" {
		config.SessionSecret = "CHANGE_THIS_IN_PRODUCTION_TO_A_SECURE_RANDOM_KEY_32_CHARS_MIN"
	}
	if config.Port == "" {
		config.Port = "8082"
	}

	return config
}