package main

import (
	"os"
	"strings"
)

// Config описывает всё, что нужно сервису выгрузки: где брать запчасти,
// где брать реквизиты продавцов и куда складывать готовый прайс-лист.
type Config struct {
	Port             string
	PartsServiceGRPC string
	AuthServiceGRPC  string
	ExportDir        string

	DromAPIURL   string
	DromAPIKey   string
	DromPacketID string

	BaseImageURL string
}

func LoadConfig() *Config {
	config := &Config{
		Port:             os.Getenv("PORT"),
		PartsServiceGRPC: os.Getenv("PARTS_SERVICE_GRPC_URL"),
		AuthServiceGRPC:  os.Getenv("AUTH_SERVICE_GRPC_URL"),
		ExportDir:        os.Getenv("EXPORT_DIR"),
		DromAPIURL:       os.Getenv("DROM_API_URL"),
		DromAPIKey:       os.Getenv("DROM_API_KEY"),
		DromPacketID:     os.Getenv("DROM_PACKET_ID"),
		BaseImageURL:     resolveBaseImageURL(),
	}

	if config.Port == "" {
		config.Port = "8085"
	}
	if config.PartsServiceGRPC == "" {
		config.PartsServiceGRPC = "parts-service:9081"
	}
	if config.AuthServiceGRPC == "" {
		config.AuthServiceGRPC = "auth-service:9083"
	}
	if config.ExportDir == "" {
		// Тот же путь, что раньше был у parts-service: адрес прайс-листа
		// снаружи не меняется, меняется только кто его отдаёт.
		config.ExportDir = "./uploads"
	}
	if config.DromAPIURL == "" {
		config.DromAPIURL = "https://baza.drom.ru/good/packet/api/sync"
	}

	return config
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
