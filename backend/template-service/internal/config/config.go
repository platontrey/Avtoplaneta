package config

import (
	"os"

	"avtoplaneta/pkg/logger"
)

type Config struct {
	Port          string
	GRPCPort      string
	DatabaseURL   string
	RedisURL      string
	RedisPassword string
	LogLevel      string
	NodeEnv       string
}

func LoadConfig() *Config {
	cfg := &Config{
		Port:          getEnv("PORT", "8090"),
		GRPCPort:      getEnv("GRPC_PORT", "9095"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/template_db?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		LogLevel:      getEnv("LOG_LEVEL", "INFO"),
		NodeEnv:       getEnv("NODE_ENV", "development"),
	}

	setupLogger(cfg.LogLevel, cfg.NodeEnv)
	return cfg
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func setupLogger(levelStr, env string) {
	fmt := logger.FormatJSON
	if env != "production" {
		fmt = logger.FormatText
	}
	logger.Init(logger.Config{
		Level:       levelStr,
		Format:      fmt,
		ServiceName: "template-service",
	})
}
