package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Port          string
	GRPCPort      string
	DatabaseURL   string
	RedisURL      string
	RedisPassword string
	NodeEnv       string
	LogLevel      string
}

// LoadConfig загружает переменные окружения с fallback-значениями
func LoadConfig() *Config {
	// Пытаемся загрузить локальный .env файл при разработке
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

	cfg := &Config{
		Port:          getEnv("PORT", "8090"),
		GRPCPort:      getEnv("GRPC_PORT", "9090"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/template_db?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		NodeEnv:       getEnv("NODE_ENV", "development"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
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
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	if env == "production" {
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "15:04:05",
		})
	}
}
