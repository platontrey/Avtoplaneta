package main

import "os"

// Config содержит конфигурационные параметры приложения
type Config struct {
	DatabaseURL    string
	SessionSecret  string
	NodeEnv        string
	Port           string
	RedisURL       string
	RedisPassword  string
	RedisDB        int
	AuthServiceURL string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	config := &Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		SessionSecret:  os.Getenv("SESSION_SECRET"),
		NodeEnv:        os.Getenv("NODE_ENV"),
		Port:           os.Getenv("PORT"),
		RedisURL:       os.Getenv("REDIS_URL"),
		RedisPassword:  os.Getenv("REDIS_PASSWORD"),
		AuthServiceURL: os.Getenv("AUTH_SERVICE_URL"),
	}

	// Значения по умолчанию
	if config.DatabaseURL == "" {
		// Fallback for local development if env var is missing, but prefer env var
		config.DatabaseURL = "host=localhost user=postgres dbname=autoplanet port=5432 sslmode=disable"
	}
	if config.SessionSecret == "" {
		config.SessionSecret = "CHANGE_THIS_IN_PRODUCTION_TO_A_SECURE_RANDOM_KEY_32_CHARS_MIN"
	}
	if config.Port == "" {
		config.Port = "8082"
	}
	if config.RedisURL == "" {
		config.RedisURL = "localhost:6379"
	}
	if config.RedisPassword == "" {
		config.RedisPassword = ""
	}
	if config.AuthServiceURL == "" {
		config.AuthServiceURL = "http://localhost:8083"
	}
	config.RedisDB = 0 // По умолчанию база данных 0

	return config
}