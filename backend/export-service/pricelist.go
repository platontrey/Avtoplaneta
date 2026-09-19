package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Прайс-лист для Drom собирается по всему складу: полная выборка запчастей плюс
// генерация XML. Раньше это происходило на каждый запрос и на каждый тик
// планировщика, даже когда склад не менялся. Теперь рядом с файлом лежит его
// отпечаток, и работа повторяется только после реальных изменений.
const (
	priceListFilename = "pricelist.xml"
	priceListMetaName = "pricelist.meta.json"
	priceListURL      = "/uploads/" + priceListFilename
)

type priceListMeta struct {
	Version    string `json:"version"`
	PartsCount int    `json:"parts_count"`
	BuiltAt    string `json:"built_at"`
}

// InventorySource — то, что сборщику нужно от остальной системы.
//
// Интерфейс объявлен здесь, у потребителя, а не рядом с *Clients: сборщику
// важно «дай отпечаток, дай запчасти, дай реквизиты», а не то, что за этим
// стоят два gRPC-соединения. Побочная польза — сборщик проверяется тестом без
// сети.
type InventorySource interface {
	InventoryVersion(ctx context.Context) (string, error)
	PartsForExport(ctx context.Context) ([]Part, error)
	SellerINNs(ctx context.Context) (map[int64]string, error)
}

// PriceListBuilder владеет файлом прайс-листа: знает, где он лежит, когда его
// нужно пересобрать и из чего собирать.
type PriceListBuilder struct {
	source InventorySource
	config *Config
}

func NewPriceListBuilder(source InventorySource, config *Config) *PriceListBuilder {
	return &PriceListBuilder{source: source, config: config}
}

func (b *PriceListBuilder) Path() string {
	return filepath.Join(b.config.ExportDir, priceListFilename)
}

func (b *PriceListBuilder) metaPath() string {
	return filepath.Join(b.config.ExportDir, priceListMetaName)
}

// Build собирает прайс-лист, если он устарел.
// Второе значение говорит, пересобирался ли файл на самом деле.
func (b *PriceListBuilder) Build(ctx context.Context) (priceListMeta, bool, error) {
	// Сначала дешёвый вопрос складу: менялось ли что-нибудь. Одна агрегирующая
	// строка вместо выкачивания всех запчастей по gRPC.
	version, err := b.source.InventoryVersion(ctx)
	if err != nil {
		// Отпечаток не прочитался — не повод отказывать в прайс-листе,
		// просто собираем его безусловно.
		version = ""
	}

	if version != "" {
		if meta, err := b.readMeta(); err == nil && meta.Version == version {
			if _, err := os.Stat(b.Path()); err == nil {
				return meta, false, nil
			}
		}
	}

	xmlData, count, err := b.Generate(ctx)
	if err != nil {
		return priceListMeta{}, false, err
	}

	if err := os.MkdirAll(b.config.ExportDir, 0o755); err != nil {
		return priceListMeta{}, false, fmt.Errorf("создать директорию выгрузки: %w", err)
	}
	if err := os.WriteFile(b.Path(), xmlData, 0o644); err != nil {
		return priceListMeta{}, false, fmt.Errorf("сохранить XML: %w", err)
	}

	meta := priceListMeta{
		Version:    version,
		PartsCount: count,
		BuiltAt:    time.Now().UTC().Format(time.RFC3339),
	}
	// Отпечаток пишется последним: если файл записался, а отпечаток нет,
	// следующий запуск просто пересоберёт прайс-лист заново.
	if payload, err := json.Marshal(meta); err == nil {
		_ = os.WriteFile(b.metaPath(), payload, 0o644)
	}

	return meta, true, nil
}

// Generate собирает XML в памяти, ничего не записывая на диск.
func (b *PriceListBuilder) Generate(ctx context.Context) ([]byte, int, error) {
	parts, err := b.source.PartsForExport(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("получить запчасти для прайс-листа: %w", err)
	}

	// ИНН — обязательный реквизит площадки, но не повод ронять выгрузку:
	// без него прайс-лист остаётся валидным, просто беднее.
	sellerINNs, err := b.source.SellerINNs(ctx)
	if err != nil {
		sellerINNs = map[int64]string{}
	}

	xmlData, err := GenerateXMLPriceList(parts, sellerINNs)
	if err != nil {
		return nil, 0, fmt.Errorf("сгенерировать XML: %w", err)
	}

	return xmlData, len(parts), nil
}

func (b *PriceListBuilder) readMeta() (priceListMeta, error) {
	payload, err := os.ReadFile(b.metaPath())
	if err != nil {
		return priceListMeta{}, err
	}
	var meta priceListMeta
	if err := json.Unmarshal(payload, &meta); err != nil {
		return priceListMeta{}, err
	}
	return meta, nil
}
