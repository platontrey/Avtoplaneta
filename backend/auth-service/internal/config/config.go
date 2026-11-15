package config

import "avtoplaneta/pkg/env"

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
	sessionSecret := env.String("SESSION_SECRET", "CHANGE_THIS_IN_PRODUCTION_TO_A_SECURE_RANDOM_KEY_32_CHARS_MIN")
	return &Config{
		DatabaseURL:     env.String("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=avtoplaneta port=5432 sslmode=disable"),
		SessionSecret:   sessionSecret,
		NodeEnv:         env.String("NODE_ENV", "development"),
		Port:            env.String("PORT", "8083"),
		AllowedOrigins:  env.String("ALLOWED_ORIGINS", ""),
		PartsServiceURL: env.String("PARTS_SERVICE_URL", "http://localhost:8081"),
		PartsGRPCAddr:   env.String("PARTS_GRPC_ADDR", "localhost:9081"),
		JWTSecret:       env.String("JWT_SECRET", sessionSecret),
		RedisURL:        env.String("REDIS_URL", ""),
	}
}
