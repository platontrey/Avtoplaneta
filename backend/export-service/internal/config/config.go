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
	// 1. Опциональный явный CDN/S3 оверрайд
	raw := strings.TrimSpace(os.Getenv("BASE_IMAGE_URL"))
	if raw == "" {
		// 2. В обычном режиме берём основной публичный домен приложения
		raw = strings.TrimSpace(os.Getenv("PUBLIC_DOMAIN"))
	}
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("DOMAIN"))
	}
	if raw == "" {
		// 3. Локальный fallback для разработки без привязки к чужим внешним доменам
		return "http://localhost:8085"
	}

	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	return strings.TrimRight(raw, "/")
}
