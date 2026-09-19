package main

import "os"

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
