package main

import "os"

// Config содержит конфигурационные параметры приложения
type Config struct {
	DatabaseURL     string
	SessionSecret   string
	NodeEnv         string
	Port            string
	AllowedOrigins  string
	PartsServiceURL string
	PartsGRPCAddr   string
	JWTSecret       string
	RedisURL        string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	config := &Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		SessionSecret:   os.Getenv("SESSION_SECRET"),
		NodeEnv:         os.Getenv("NODE_ENV"),
		Port:            os.Getenv("PORT"),
		AllowedOrigins:  os.Getenv("ALLOWED_ORIGINS"),
		PartsServiceURL: os.Getenv("PARTS_SERVICE_URL"),
		PartsGRPCAddr:   os.Getenv("PARTS_GRPC_ADDR"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		RedisURL:        os.Getenv("REDIS_URL"),
	}

	// Значения по умолчанию
	if config.PartsGRPCAddr == "" {
		config.PartsGRPCAddr = "localhost:9081"
	}
	if config.PartsServiceURL == "" {
		config.PartsServiceURL = "http://localhost:8081"
	}
	if config.DatabaseURL == "" {
		config.DatabaseURL = "host=localhost user=postgres password=qewret123 dbname=autoplanet port=5432 sslmode=disable"
	}
	if config.SessionSecret == "" {
		config.SessionSecret = "CHANGE_THIS_IN_PRODUCTION_TO_A_SECURE_RANDOM_KEY_32_CHARS_MIN"
	}
	if config.Port == "" {
		config.Port = "8083"
	}
	if config.JWTSecret == "" {
		config.JWTSecret = config.SessionSecret // fallback — используем SESSION_SECRET
	}

	return config
}