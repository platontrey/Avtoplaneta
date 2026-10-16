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
	GetPart(ctx context.Context, partID int64) (*Part, error)
	SellerINNs(ctx context.Context) (map[int64]string, error)
}

// PriceListBuilder владеет файлом прайс-листа и локальным слепком каталога:
// знает, где он лежит, когда его нужно пересобрать и из чего собирать.
type PriceListBuilder struct {
	source InventorySource
	store  *SnapshotStore
	config *Config
}

func NewPriceListBuilder(source InventorySource, store *SnapshotStore, config *Config) *PriceListBuilder {
	return &PriceListBuilder{source: source, store: store, config: config}
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

	if b.store != nil {
		storeCount := b.store.Count()
		storeVersion := b.store.GetVersion()

		// 1. Слепок актуален и файл на диске существует -> пересборка не требуется
		if storeCount > 0 && version != "" && storeVersion == version {
			if _, err := os.Stat(b.Path()); err == nil {
				meta := priceListMeta{
					Version:    version,
					PartsCount: storeCount,
					BuiltAt:    time.Now().UTC().Format(time.RFC3339),
				}
				return meta, false, nil
			}
			// Файл на диске случайно удален, но слепок в bbolt жив -> мгновенно сбрасываем без сети
			count, writeErr := b.store.WriteFullXML(b.Path())
			if writeErr == nil {
				meta := priceListMeta{
					Version:    version,
					PartsCount: count,
					BuiltAt:    time.Now().UTC().Format(time.RFC3339),
				}
				b.writeMeta(meta)
				return meta, true, nil
			}
		}

		// 2. Если слепок пуст или склад изменился без поштучных событий -> первоначальная загрузка слепка
		if storeCount == 0 || (version != "" && storeVersion != version) {
			parts, err := b.source.PartsForExport(ctx)
			if err != nil {
				return priceListMeta{}, false, fmt.Errorf("получить запчасти для слепка: %w", err)
			}
			sellerINNs, err := b.source.SellerINNs(ctx)
			if err != nil {
				sellerINNs = map[int64]string{}
			}

			if err := b.store.BulkLoad(parts, sellerINNs, b.config.BaseImageURL, version); err != nil {
				return priceListMeta{}, false, fmt.Errorf("загрузить каталог в слепок: %w", err)
			}

			count, err := b.store.WriteFullXML(b.Path())
			if err != nil {
				return priceListMeta{}, false, fmt.Errorf("сохранить XML из слепка: %w", err)
			}

			meta := priceListMeta{
				Version:    version,
				PartsCount: count,
				BuiltAt:    time.Now().UTC().Format(time.RFC3339),
			}
			b.writeMeta(meta)
			return meta, true, nil
		}
	}

	// Резервный путь при работе без bbolt (например в легких юнит-тестах)
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
	b.writeMeta(meta)

	return meta, true, nil
}

// UpdatePartInSnapshot точечно запрашивает одну деталь и обновляет её XML-кусочек в слепке
func (b *PriceListBuilder) UpdatePartInSnapshot(ctx context.Context, partID int64) error {
	if b.store == nil {
		return nil
	}

	part, err := b.source.GetPart(ctx, partID)
	if err != nil {
		return fmt.Errorf("запросить обновленную деталь %d: %w", partID, err)
	}

	sellerINNs, err := b.source.SellerINNs(ctx)
	if err != nil {
		sellerINNs = map[int64]string{}
	}

	if err := b.store.UpsertOffer(*part, sellerINNs[part.SellerID], b.config.BaseImageURL); err != nil {
		return fmt.Errorf("обновить оффер в слепке: %w", err)
	}

	// Атомарно обновляем XML на диске из готовых байт слепка (~30 мс)
	count, err := b.store.WriteFullXML(b.Path())
	if err != nil {
		return fmt.Errorf("записать XML после обновления детали: %w", err)
	}

	meta := priceListMeta{
		Version:    b.store.GetVersion(),
		PartsCount: count,
		BuiltAt:    time.Now().UTC().Format(time.RFC3339),
	}
	b.writeMeta(meta)
	return nil
}

// DeletePartFromSnapshot удаляет деталь из локального слепка и мгновенно перезаписывает XML
func (b *PriceListBuilder) DeletePartFromSnapshot(ctx context.Context, partID int64) error {
	if b.store == nil {
		return nil
	}

	if err := b.store.DeleteOffer(partID); err != nil {
		return fmt.Errorf("удалить оффер из слепка: %w", err)
	}

	count, err := b.store.WriteFullXML(b.Path())
	if err != nil {
		return fmt.Errorf("записать XML после удаления детали: %w", err)
	}

	meta := priceListMeta{
		Version:    b.store.GetVersion(),
		PartsCount: count,
		BuiltAt:    time.Now().UTC().Format(time.RFC3339),
	}
	b.writeMeta(meta)
	return nil
}

// Generate собирает XML в памяти, ничего не записывая на диск.
func (b *PriceListBuilder) Generate(ctx context.Context) ([]byte, int, error) {
	if b.store != nil && b.store.Count() > 0 {
		return b.store.GenerateFullXMLBytes()
	}

	parts, err := b.source.PartsForExport(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("получить запчасти для прайс-листа: %w", err)
	}

	sellerINNs, err := b.source.SellerINNs(ctx)
	if err != nil {
		sellerINNs = map[int64]string{}
	}

	xmlData, err := GenerateXMLPriceList(parts, sellerINNs, b.config.BaseImageURL)
	if err != nil {
		return nil, 0, fmt.Errorf("сгенерировать XML: %w", err)
	}

	return xmlData, len(parts), nil
}

func (b *PriceListBuilder) writeMeta(meta priceListMeta) {
	if payload, err := json.Marshal(meta); err == nil {
		_ = os.WriteFile(b.metaPath(), payload, 0o644)
	}
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
