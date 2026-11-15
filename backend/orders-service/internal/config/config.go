package config

import (
	"avtoplaneta/pkg/env"
)

// Config содержит конфигурационные параметры приложения orders-service
type Config struct {
	DatabaseURL         string
	SessionSecret       string
	NodeEnv             string
	Port                string
	GRPCPort            string
	RedisURL            string
	RedisPassword       string
	RedisDB             int
	AuthServiceURL      string
	AuthServiceGRPCURL  string
	PartsServiceGRPCURL string
	LogLevel            string
	LogFormat           string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	defaultLogFormat := "text"
	if env.String("NODE_ENV", "") == "production" {
		defaultLogFormat = "json"
	}

	return &Config{
		DatabaseURL:         env.FirstNonEmpty([]string{"ORDERS_DATABASE_URL", "DATABASE_URL"}, "host=localhost user=postgres dbname=autoplanet port=5432 sslmode=disable"),
		SessionSecret:       env.String("SESSION_SECRET", "CHANGE_THIS_IN_PRODUCTION_TO_A_SECURE_RANDOM_KEY_32_CHARS_MIN"),
		NodeEnv:             env.String("NODE_ENV", "development"),
		Port:                env.String("PORT", "8082"),
		GRPCPort:            env.String("GRPC_PORT", "9082"),
		RedisURL:            env.String("REDIS_URL", "localhost:6379"),
		RedisPassword:       env.String("REDIS_PASSWORD", ""),
		RedisDB:             env.Int("REDIS_DB", 0),
		AuthServiceURL:      env.String("AUTH_SERVICE_URL", "http://localhost:8083"),
		AuthServiceGRPCURL:  env.FirstNonEmpty([]string{"AUTH_SERVICE_GRPC_URL", "AUTH_GRPC_ADDR"}, "localhost:9083"),
		PartsServiceGRPCURL: env.String("PARTS_SERVICE_GRPC_URL", "localhost:9081"),
		LogLevel:            env.String("LOG_LEVEL", "info"),
		LogFormat:           env.String("LOG_FORMAT", defaultLogFormat),
	}
}
