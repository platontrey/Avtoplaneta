package config

import (
	"os"
	"strings"
)

// Config описывает параметры окружения сервиса экспорта каталогов
type Config struct {
	Port             string
	PartsServiceGRPC string
	AuthServiceGRPC  string
	ExportDir        string
	CatalogDBPath    string

	RedisURL      string
	RedisPassword string

	DromAPIURL   string
	DromAPIKey   string
	DromPacketID string

	BaseImageURL string
}

// LoadConfig читает конфигурацию из переменных окружения с разумными дефолтами
func LoadConfig() *Config {
	exportDir := os.Getenv("EXPORT_DIR")
	if exportDir == "" {
		exportDir = "./uploads"
	}

	catalogDBPath := os.Getenv("CATALOG_DB_PATH")
	if catalogDBPath == "" {
		catalogDBPath = exportDir + "/catalog.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	partsGRPC := os.Getenv("PARTS_SERVICE_GRPC_URL")
	if partsGRPC == "" {
		partsGRPC = "parts-service:9081"
	}

	authGRPC := os.Getenv("AUTH_SERVICE_GRPC_URL")
	if authGRPC == "" {
		authGRPC = "auth-service:9083"
	}

	dromURL := os.Getenv("DROM_API_URL")
	if dromURL == "" {
		dromURL = "https://baza.drom.ru/good/packet/api/sync"
	}

	return &Config{
		Port:             port,
		PartsServiceGRPC: partsGRPC,
		AuthServiceGRPC:  authGRPC,
		ExportDir:        exportDir,
		CatalogDBPath:    catalogDBPath,
		RedisURL:         os.Getenv("REDIS_URL"),
		RedisPassword:    os.Getenv("REDIS_PASSWORD"),
		DromAPIURL:       dromURL,
		DromAPIKey:       os.Getenv("DROM_API_KEY"),
		DromPacketID:     os.Getenv("DROM_PACKET_ID"),
		BaseImageURL:     resolveBaseImageURL(),
	}
}

func resolveBaseImageURL() string {
	raw := os.Getenv("BASE_IMAGE_URL")
	if raw == "" {
		raw = os.Getenv("PUBLIC_DOMAIN")
	}
	if raw == "" {
		raw = os.Getenv("DOMAIN")
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "https://backend-server.ru"
	}

	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	return strings.TrimRight(raw, "/")
}
