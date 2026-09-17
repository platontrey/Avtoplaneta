package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Прайс-лист для Drom собирается по всему складу: полная выборка запчастей плюс
// генерация XML. Раньше это происходило на каждый запрос и на каждый тик
// планировщика, даже когда склад не менялся. Теперь рядом с файлом лежит его
// отпечаток, и работа повторяется только после реальных изменений.
const (
	priceListDir      = "./uploads"
	priceListFilename = "pricelist.xml"
	priceListPath     = priceListDir + "/" + priceListFilename
	priceListMetaPath = priceListDir + "/pricelist.meta.json"
	priceListURL      = "/uploads/" + priceListFilename
)

type priceListMeta struct {
	Version    string `json:"version"`
	PartsCount int    `json:"parts_count"`
	BuiltAt    string `json:"built_at"`
}

// buildPriceList собирает прайс-лист, если он устарел.
// Второе значение говорит, пересобирался ли файл на самом деле.
func buildPriceList(ctx context.Context) (priceListMeta, bool, error) {
	repo := NewPartRepository(dbPool)
	version, err := repo.InventoryVersion(ctx)
	if err != nil {
		// Отпечаток не прочитался — не повод отказывать в прайс-листе,
		// просто собираем его безусловно.
		version = ""
	}

	if version != "" {
		if meta, err := readPriceListMeta(); err == nil && meta.Version == version {
			if _, err := os.Stat(priceListPath); err == nil {
				return meta, false, nil
			}
		}
	}

	parts, err := GetPartsForXML()
	if err != nil {
		return priceListMeta{}, false, fmt.Errorf("получить запчасти для прайс-листа: %w", err)
	}

	xmlData, err := GenerateXMLPriceList(parts)
	if err != nil {
		return priceListMeta{}, false, fmt.Errorf("сгенерировать XML: %w", err)
	}

	if err := os.MkdirAll(priceListDir, 0o755); err != nil {
		return priceListMeta{}, false, fmt.Errorf("создать директорию uploads: %w", err)
	}
	if err := os.WriteFile(priceListPath, xmlData, 0o644); err != nil {
		return priceListMeta{}, false, fmt.Errorf("сохранить XML: %w", err)
	}

	meta := priceListMeta{
		Version:    version,
		PartsCount: len(parts),
		BuiltAt:    time.Now().UTC().Format(time.RFC3339),
	}
	// Отпечаток пишется последним: если файл записался, а отпечаток нет,
	// следующий запуск просто пересоберёт прайс-лист заново.
	if payload, err := json.Marshal(meta); err == nil {
		_ = os.WriteFile(priceListMetaPath, payload, 0o644)
	}

	return meta, true, nil
}

func readPriceListMeta() (priceListMeta, error) {
	payload, err := os.ReadFile(priceListMetaPath)
	if err != nil {
		return priceListMeta{}, err
	}
	var meta priceListMeta
	if err := json.Unmarshal(payload, &meta); err != nil {
		return priceListMeta{}, err
	}
	return meta, nil
}
