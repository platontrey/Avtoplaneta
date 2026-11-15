package config

import (
	"avtoplaneta/pkg/env"
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
		Port:          env.String("PORT", "8090"),
		GRPCPort:      env.String("GRPC_PORT", "9095"),
		DatabaseURL:   env.String("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/template_db?sslmode=disable"),
		RedisURL:      env.String("REDIS_URL", "localhost:6379"),
		RedisPassword: env.String("REDIS_PASSWORD", ""),
		LogLevel:      env.String("LOG_LEVEL", "INFO"),
		NodeEnv:       env.String("NODE_ENV", "development"),
	}

	setupLogger(cfg.LogLevel, cfg.NodeEnv)
	return cfg
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
