package config

import "os"

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
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	dbURL := os.Getenv("ORDERS_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		dbURL = "host=localhost user=postgres dbname=autoplanet port=5432 sslmode=disable"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9082"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "CHANGE_THIS_IN_PRODUCTION_TO_A_SECURE_RANDOM_KEY_32_CHARS_MIN"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	authURL := os.Getenv("AUTH_SERVICE_URL")
	if authURL == "" {
		authURL = "http://localhost:8083"
	}

	authGRPCURL := os.Getenv("AUTH_SERVICE_GRPC_URL")
	if authGRPCURL == "" {
		if addr := os.Getenv("AUTH_GRPC_ADDR"); addr != "" {
			authGRPCURL = addr
		} else {
			authGRPCURL = "localhost:9083"
		}
	}

	partsGRPCURL := os.Getenv("PARTS_SERVICE_GRPC_URL")
	if partsGRPCURL == "" {
		partsGRPCURL = "localhost:9081"
	}

	return &Config{
		DatabaseURL:         dbURL,
		SessionSecret:       sessionSecret,
		NodeEnv:             os.Getenv("NODE_ENV"),
		Port:                port,
		GRPCPort:            grpcPort,
		RedisURL:            redisURL,
		RedisPassword:       os.Getenv("REDIS_PASSWORD"),
		RedisDB:             0,
		AuthServiceURL:      authURL,
		AuthServiceGRPCURL:  authGRPCURL,
		PartsServiceGRPCURL: partsGRPCURL,
	}
}
