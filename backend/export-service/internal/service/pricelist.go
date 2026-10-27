package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"export-service/internal/config"
	"export-service/internal/domain"
	"export-service/internal/repository"
)

const (
	// PriceListFilename имя файла выгрузки каталога
	PriceListFilename = "pricelist.xml"
	// PriceListMetaName имя файла метаданных последней сборки
	PriceListMetaName = "pricelist.meta.json"
	// PriceListURL публичный URL для скачивания файла
	PriceListURL = "/uploads/" + PriceListFilename
)

// PriceListService управляет жизненным циклом и сборкой XML прайс-листа для Drom.
//
// Сервис использует инкрементальный слепок SnapshotRepository (на базе bbolt),
// позволяя мгновенно собирать 50 000+ офферов без обращения к базе данных parts-service.
type PriceListService struct {
	source InventorySource
	repo   repository.SnapshotRepository
	config *config.Config
}

// NewPriceListService создает экземпляр PriceListService
func NewPriceListService(source InventorySource, repo repository.SnapshotRepository, cfg *config.Config) *PriceListService {
	return &PriceListService{
		source: source,
		repo:   repo,
		config: cfg,
	}
}

// Path возвращает абсолютный путь к целевому XML файлу
func (s *PriceListService) Path() string {
	return filepath.Join(s.config.ExportDir, PriceListFilename)
}

// MetaPath возвращает путь к JSON файлу метаданных
func (s *PriceListService) MetaPath() string {
	return filepath.Join(s.config.ExportDir, PriceListMetaName)
}

// Build выполняет сборку XML прайс-листа только при реальных изменениях каталога.
// Возвращает метаданные, флаг факта пересборки и ошибку.
func (s *PriceListService) Build(ctx context.Context) (domain.PriceListMeta, bool, error) {
	// Сначала запрашиваем отпечаток склада (дешевый агрегирующий запрос)
	version, err := s.source.InventoryVersion(ctx)
	if err != nil {
		version = ""
	}

	if s.repo != nil {
		repoCount := s.repo.Count()
		repoVersion := s.repo.GetVersion()

		// 1. Слепок актуален и файл на диске существует -> пересборка не требуется
		if repoCount > 0 && version != "" && repoVersion == version {
			if _, err := os.Stat(s.Path()); err == nil {
				meta := domain.PriceListMeta{
					Version:    version,
					PartsCount: repoCount,
					BuiltAt:    time.Now().UTC().Format(time.RFC3339),
				}
				return meta, false, nil
			}
			// Файл на диске отсутствует, но слепок актуален -> мгновенный сброс без gRPC
			count, writeErr := s.repo.WriteFullXML(s.Path())
			if writeErr == nil {
				meta := domain.PriceListMeta{
					Version:    version,
					PartsCount: count,
					BuiltAt:    time.Now().UTC().Format(time.RFC3339),
				}
				s.writeMeta(meta)
				return meta, true, nil
			}
		}

		// 2. Если слепок пуст или склад изменился -> полная первоначальная загрузка слепка
		if repoCount == 0 || (version != "" && repoVersion != version) {
			parts, err := s.source.PartsForExport(ctx)
			if err != nil {
				return domain.PriceListMeta{}, false, fmt.Errorf("получить запчасти для слепка: %w", err)
			}
			sellerINNs, err := s.source.SellerINNs(ctx)
			if err != nil {
				sellerINNs = map[int64]string{}
			}

			if err := s.repo.BulkLoad(parts, sellerINNs, s.config.BaseImageURL, version); err != nil {
				return domain.PriceListMeta{}, false, fmt.Errorf("загрузить каталог в слепок: %w", err)
			}

			count, err := s.repo.WriteFullXML(s.Path())
			if err != nil {
				return domain.PriceListMeta{}, false, fmt.Errorf("сохранить XML из слепка: %w", err)
			}

			meta := domain.PriceListMeta{
				Version:    version,
				PartsCount: count,
				BuiltAt:    time.Now().UTC().Format(time.RFC3339),
			}
			s.writeMeta(meta)
			return meta, true, nil
		}
	}

	// Резервный путь без репозитория слепка (например, в минимальных юнит-тестах)
	if version != "" {
		if meta, err := s.readMeta(); err == nil && meta.Version == version {
			if _, err := os.Stat(s.Path()); err == nil {
				return meta, false, nil
			}
		}
	}

	xmlData, count, err := s.Generate(ctx)
	if err != nil {
		return domain.PriceListMeta{}, false, err
	}

	if err := os.MkdirAll(s.config.ExportDir, 0o755); err != nil {
		return domain.PriceListMeta{}, false, fmt.Errorf("создать директорию выгрузки: %w", err)
	}
	if err := os.WriteFile(s.Path(), xmlData, 0o644); err != nil {
		return domain.PriceListMeta{}, false, fmt.Errorf("сохранить XML: %w", err)
	}

	meta := domain.PriceListMeta{
		Version:    version,
		PartsCount: count,
		BuiltAt:    time.Now().UTC().Format(time.RFC3339),
	}
	s.writeMeta(meta)

	return meta, true, nil
}

// UpdatePartInSnapshot точечно запрашивает одну деталь и обновляет её оффер в слепке
func (s *PriceListService) UpdatePartInSnapshot(ctx context.Context, partID int64) error {
	if s.repo == nil {
		return nil
	}

	part, err := s.source.GetPart(ctx, partID)
	if err != nil {
		return fmt.Errorf("запросить обновленную деталь %d: %w", partID, err)
	}

	sellerINNs, err := s.source.SellerINNs(ctx)
	if err != nil {
		sellerINNs = map[int64]string{}
	}

	if err := s.repo.SaveOffer(*part, sellerINNs[part.SellerID], s.config.BaseImageURL); err != nil {
		return fmt.Errorf("обновить оффер в слепке: %w", err)
	}

	count, err := s.repo.WriteFullXML(s.Path())
	if err != nil {
		return fmt.Errorf("записать XML после обновления детали: %w", err)
	}

	meta := domain.PriceListMeta{
		Version:    s.repo.GetVersion(),
		PartsCount: count,
		BuiltAt:    time.Now().UTC().Format(time.RFC3339),
	}
	s.writeMeta(meta)
	return nil
}

// DeletePartFromSnapshot удаляет деталь из локального слепка и перезаписывает XML
func (s *PriceListService) DeletePartFromSnapshot(ctx context.Context, partID int64) error {
	if s.repo == nil {
		return nil
	}

	if err := s.repo.DeleteOffer(partID); err != nil {
		return fmt.Errorf("удалить оффер из слепка: %w", err)
	}

	count, err := s.repo.WriteFullXML(s.Path())
	if err != nil {
		return fmt.Errorf("записать XML после удаления детали: %w", err)
	}

	meta := domain.PriceListMeta{
		Version:    s.repo.GetVersion(),
		PartsCount: count,
		BuiltAt:    time.Now().UTC().Format(time.RFC3339),
	}
	s.writeMeta(meta)
	return nil
}

// Generate собирает XML в памяти без записи на диск
func (s *PriceListService) Generate(ctx context.Context) ([]byte, int, error) {
	if s.repo != nil && s.repo.Count() > 0 {
		return s.repo.GenerateFullXMLBytes()
	}

	parts, err := s.source.PartsForExport(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("получить запчасти для прайс-листа: %w", err)
	}

	sellerINNs, err := s.source.SellerINNs(ctx)
	if err != nil {
		sellerINNs = map[int64]string{}
	}

	xmlData, err := GenerateXMLPriceList(parts, sellerINNs, s.config.BaseImageURL)
	if err != nil {
		return nil, 0, fmt.Errorf("сгенерировать XML: %w", err)
	}

	return xmlData, len(parts), nil
}

func (s *PriceListService) writeMeta(meta domain.PriceListMeta) {
	if payload, err := json.Marshal(meta); err == nil {
		_ = os.WriteFile(s.MetaPath(), payload, 0o644)
	}
}

func (s *PriceListService) readMeta() (domain.PriceListMeta, error) {
	payload, err := os.ReadFile(s.MetaPath())
	if err != nil {
		return domain.PriceListMeta{}, err
	}
	var meta domain.PriceListMeta
	if err := json.Unmarshal(payload, &meta); err != nil {
		return domain.PriceListMeta{}, err
	}
	return meta, nil
}
