package config

import (
	"avtoplaneta/pkg/env"
)

// Config содержит конфигурационные параметры приложения
type Config struct {
	DatabaseURL      string
	Port             string
	RedisURL         string
	RedisPassword    string
	ElasticsearchURL string
	LogLevel         string
	LogFormat        string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	defaultLogFormat := "text"
	if env.String("NODE_ENV", "") == "production" {
		defaultLogFormat = "json"
	}

	return &Config{
		DatabaseURL:      env.String("DATABASE_URL", "host=localhost user=postgres dbname=autoplanet port=5432 sslmode=disable"),
		Port:             env.String("PORT", "8081"),
		RedisURL:         env.String("REDIS_URL", "127.0.0.1:6379"),
		RedisPassword:    env.String("REDIS_PASSWORD", ""),
		ElasticsearchURL: env.String("ELASTICSEARCH_URL", "http://localhost:9200"),
		LogLevel:         env.String("LOG_LEVEL", "info"),
		LogFormat:        env.String("LOG_FORMAT", defaultLogFormat),
	}
}
