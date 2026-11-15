package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"avtoplaneta/pkg/logger"
	"parts-service/internal/domain"
	"parts-service/internal/search"
)

// DeleteZeroQuantityPartsBySupplier удаляет запчасти с нулевым количеством по поставщику
func (s *inventoryService) DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) {
	fmt.Printf("Service: DeleteZeroQuantityPartsBySupplier called with supplier_code='%s'\n", supplierCode)

	var (
		deletedCount int64
		deletedIDs   []int64
		err          error
	)

	if idDeleter, ok := s.repo.(interface {
		DeleteZeroQuantityPartsWithIDs(ctx context.Context, supplierCode string) ([]int64, error)
	}); ok {
		deletedIDs, err = idDeleter.DeleteZeroQuantityPartsWithIDs(ctx, supplierCode)
		deletedCount = int64(len(deletedIDs))
	} else {
		deletedCount, err = s.repo.DeleteZeroQuantityPartsBySupplier(ctx, supplierCode)
	}
	if err != nil {
		fmt.Printf("Service: DeleteZeroQuantityPartsBySupplier failed for supplier_code='%s': %v\n", supplierCode, err)
		return 0, err
	}

	if len(deletedIDs) > 0 {
		if esErr := search.BulkDeletePartsFromIndex(ctx, deletedIDs); esErr != nil {
			slog.WarnContext(ctx, "Failed to bulk delete zero-quantity parts from Elasticsearch", logger.Err(esErr))
		}
	}

	if s.redis != nil {
		s.redis.Del(ctx, inventoryCacheKey)
		s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
		s.invalidateStatisticsCache(ctx)
	}

	fmt.Printf("Service: DeleteZeroQuantityPartsBySupplier completed successfully for supplier_code='%s', deleted %d parts\n", supplierCode, deletedCount)
	return deletedCount, nil
}

// GetSupplierBatches получает расширенную информацию о поставках / ведомостях с нулевым количеством
func (s *inventoryService) GetSupplierBatches(ctx context.Context) ([]domain.SupplierBatchInfo, error) {
	if batchProvider, ok := s.repo.(interface {
		GetSupplierBatches(ctx context.Context) ([]domain.SupplierBatchInfo, error)
	}); ok {
		return batchProvider.GetSupplierBatches(ctx)
	}
	codes, err := s.repo.GetSupplierCodes(ctx)
	if err != nil {
		return nil, err
	}
	batches := make([]domain.SupplierBatchInfo, 0, len(codes))
	for _, code := range codes {
		batches = append(batches, domain.SupplierBatchInfo{Code: code, Label: code})
	}
	return batches, nil
}

// GetSupplierCodes получает коды поставщиков
func (s *inventoryService) GetSupplierCodes(ctx context.Context) ([]string, error) {
	codes, err := s.repo.GetSupplierCodes(ctx)
	if err != nil {
		fmt.Printf("Service: GetSupplierCodes failed: %v\n", err)
		return nil, err
	}
	fmt.Printf("Service: GetSupplierCodes returned %d codes\n", len(codes))
	return codes, nil
}

// GetNextSupplierCode получает следующий номер поставки через репозиторий
func (s *inventoryService) GetNextSupplierCode(ctx context.Context) (string, error) {
	return s.repo.GetNextSupplierCode(ctx)
}

// PeekNextSupplierCode просматривает следующий номер поставки без инкремента
func (s *inventoryService) PeekNextSupplierCode(ctx context.Context) (string, error) {
	return s.repo.PeekNextSupplierCode(ctx)
}

// GetShelves возвращает список полок склада (с кэшированием в Redis)
func (s *inventoryService) GetShelves(ctx context.Context) ([]domain.Shelf, error) {
	const cacheKey = "parts:shelves:list"
	if s.redis != nil {
		if cached, err := s.redis.Get(ctx, cacheKey).Bytes(); err == nil && len(cached) > 0 {
			var shelves []domain.Shelf
			if err := json.Unmarshal(cached, &shelves); err == nil {
				return shelves, nil
			}
		}
	}

	shelves, err := s.repo.GetShelves(ctx)
	if err != nil {
		return nil, err
	}

	if s.redis != nil && len(shelves) > 0 {
		if data, err := json.Marshal(shelves); err == nil {
			_ = s.redis.Set(ctx, cacheKey, data, 10*time.Minute).Err()
		}
	}
	return shelves, nil
}

// CreateShelf добавляет новую полку склада
func (s *inventoryService) CreateShelf(ctx context.Context, name string) (*domain.Shelf, error) {
	shelf, err := s.repo.CreateShelf(ctx, name)
	if err != nil {
		return nil, err
	}
	if s.redis != nil {
		_ = s.redis.Del(ctx, "parts:shelves:list").Err()
	}
	return shelf, nil
}

// DeleteShelf удаляет полку склада по ID
func (s *inventoryService) DeleteShelf(ctx context.Context, id int64) error {
	if err := s.repo.DeleteShelf(ctx, id); err != nil {
		return err
	}
	if s.redis != nil {
		_ = s.redis.Del(ctx, "parts:shelves:list").Err()
	}
	return nil
}
