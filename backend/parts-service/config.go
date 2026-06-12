package main

import "os"

// Config содержит конфигурационные параметры приложения
type Config struct {
	DatabaseURL      string
	Port             string
	RedisURL         string
	RedisPassword    string
	ElasticsearchURL string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	config := &Config{
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		Port:             os.Getenv("PORT"),
		RedisURL:         os.Getenv("REDIS_URL"),
		RedisPassword:    os.Getenv("REDIS_PASSWORD"),
		ElasticsearchURL: os.Getenv("ELASTICSEARCH_URL"),
	}

	// Значения по умолчанию
	if config.DatabaseURL == "" {
		config.DatabaseURL = "host=localhost user=postgres dbname=autoplanet port=5432 sslmode=disable"
	}
	if config.Port == "" {
		config.Port = "8081"
	}
	if config.RedisURL == "" {
		config.RedisURL = "127.0.0.1:6379"
	}
	if config.RedisPassword == "" {
		config.RedisPassword = ""
	}
	if config.ElasticsearchURL == "" {
		config.ElasticsearchURL = "http://localhost:9200"
	}

	return config
}