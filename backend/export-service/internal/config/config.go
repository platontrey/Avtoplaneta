package config

import (
	"strings"

	"avtoplaneta/pkg/env"
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
	exportDir := env.String("EXPORT_DIR", "./uploads")
	return &Config{
		Port:             env.String("PORT", "8085"),
		PartsServiceGRPC: env.String("PARTS_SERVICE_GRPC_URL", "parts-service:9081"),
		AuthServiceGRPC:  env.String("AUTH_SERVICE_GRPC_URL", "auth-service:9083"),
		ExportDir:        exportDir,
		CatalogDBPath:    env.String("CATALOG_DB_PATH", exportDir+"/catalog.db"),
		RedisURL:         env.String("REDIS_URL", ""),
		RedisPassword:    env.String("REDIS_PASSWORD", ""),
		DromAPIURL:       env.String("DROM_API_URL", "https://baza.drom.ru/good/packet/api/sync"),
		DromAPIKey:       env.String("DROM_API_KEY", ""),
		DromPacketID:     env.String("DROM_PACKET_ID", ""),
		BaseImageURL:     resolveBaseImageURL(),
	}
}

func resolveBaseImageURL() string {
	raw := env.FirstNonEmpty([]string{"BASE_IMAGE_URL", "PUBLIC_DOMAIN", "DOMAIN"}, "http://localhost:8085")
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	return strings.TrimRight(raw, "/")
}
