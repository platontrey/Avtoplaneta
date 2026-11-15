package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	partsv1 "avtoplaneta/gen/parts/v1"
	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/logger"
	"parts-service/internal/domain"
	"parts-service/internal/search"
)

// ErrEntireCatalogConfirmationRequired возвращается, когда массовая операция по пустому фильтру
// затрагивает весь каталог, а клиент не передал confirm_entire_catalog.
var ErrEntireCatalogConfirmationRequired = errors.New("массовая операция над всем каталогом требует явного подтверждения (confirm_entire_catalog)")

func isFilterEmpty(f *partsv1.GetInventoryRequest) bool {
	if f == nil {
		return true
	}
	hasPhotoFilter := f.HasPhoto != "" && f.HasPhoto != "all"
	return !hasPhotoFilter &&
		f.Search == "" && f.Category == "" && f.Brand == "" && f.Model == "" &&
		f.Location == "" && f.Salesman == "" && f.Status == "" && f.Address == "" &&
		f.Number == "" && f.OemCode == "" && f.Vin == "" && f.BodyBrand == "" &&
		f.EngineBrand == "" && f.CarReleaseDate == "" && f.CarReleasePeriod == "" &&
		f.Transmission == "" && f.Drive == "" && f.Condition == "" &&
		f.Manufacturer == "" && f.Defect == "" && f.Color == "" &&
		f.MinPrice == "" && f.MaxPrice == "" && f.MinQuantity == "" && f.MaxQuantity == "" &&
		f.FrontRear == "" && f.LeftRight == "" && f.TopBottom == "" &&
		f.ManufacturerCode == "" && f.SupplierCode == "" && f.TransmissionModel == "" &&
		f.WearPercentage == "" && f.Season == "" && f.Diameter == "" &&
		f.Width == "" && f.Profile == "" && f.TireQuantity == "" && f.Drilling == "" &&
		f.Offset == "" && f.CenterHoleDiameter == "" && f.TireModel == ""
}

func inventoryQueryParamsFromProtoFilter(f *partsv1.GetInventoryRequest) domain.InventoryQueryParams {
	if f == nil {
		return domain.InventoryQueryParams{}
	}
	return domain.InventoryQueryParams{
		Search:             f.Search,
		Category:           f.Category,
		Brand:              f.Brand,
		Model:              f.Model,
		Location:           f.Location,
		Address:            f.Address,
		Salesman:           f.Salesman,
		Status:             f.Status,
		HasPhoto:           f.HasPhoto,
		Number:             f.Number,
		OEMCode:            f.OemCode,
		VIN:                f.Vin,
		BodyBrand:          f.BodyBrand,
		EngineBrand:        f.EngineBrand,
		CarReleaseDate:     f.CarReleaseDate,
		CarReleasePeriod:   f.CarReleasePeriod,
		Transmission:       f.Transmission,
		Drive:              f.Drive,
		Condition:          f.Condition,
		Manufacturer:       f.Manufacturer,
		Defect:             f.Defect,
		Color:              f.Color,
		MinPrice:           f.MinPrice,
		MaxPrice:           f.MaxPrice,
		MinQuantity:        f.MinQuantity,
		MaxQuantity:        f.MaxQuantity,
		FrontRear:          f.FrontRear,
		LeftRight:          f.LeftRight,
		TopBottom:          f.TopBottom,
		ManufacturerCode:   f.ManufacturerCode,
		SupplierCode:       f.SupplierCode,
		TransmissionModel:  f.TransmissionModel,
		WearPercentage:     f.WearPercentage,
		Season:             f.Season,
		Diameter:           f.Diameter,
		Width:              f.Width,
		Profile:            f.Profile,
		TireQuantity:       f.TireQuantity,
		Drilling:           f.Drilling,
		Offset:             f.Offset,
		CenterHoleDiameter: f.CenterHoleDiameter,
		TireModel:          f.TireModel,
		Limit:              10000,
		Page:               1,
	}
}

func (s *inventoryService) searchPartIDsViaES(ctx context.Context, params domain.InventoryQueryParams) ([]int64, bool) {
	if strings.TrimSpace(params.Search) == "" || s.es == nil {
		return nil, false
	}
	esQuery := s.buildElasticsearchQuery(params)
	esParts, total, err := s.es.SearchParts(esQuery, 0, 10000)
	if err != nil {
		slog.WarnContext(ctx, "resolvePartIDsFromFilter: Failed to search via Elasticsearch, falling back to database", logger.Err(err))
		return nil, false
	}
	if total > 10000 {
		slog.InfoContext(ctx, "resolvePartIDsFromFilter: Search matched >10000 items in Elasticsearch, resolving all IDs via database", slog.Int64("total", total))
		return nil, false
	}
	ids := make([]int64, len(esParts))
	for i, p := range esParts {
		ids[i] = p.ID
	}
	return ids, true
}

func (s *inventoryService) resolvePartIDsFromFilter(ctx context.Context, f *partsv1.GetInventoryRequest) ([]int64, error) {
	params := inventoryQueryParamsFromProtoFilter(f)

	// Полнотекстовый поиск через Elasticsearch используем только при наличии поисковой строки Search
	if ids, ok := s.searchPartIDsViaES(ctx, params); ok {
		return ids, nil
	}

	filterMap := s.buildDatabaseFilters(params)
	filterMap["to_delete_at_is_null"] = true
	filterMap["quantity_gte"] = 0

	// limit = 0 означает отсутствие лимита в FindWithFilters (выбираются все совпавшие записи)
	dbParts, err := s.repo.FindWithFilters(ctx, filterMap, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("resolvePartIDsFromFilter: database query failed: %w", err)
	}
	ids := make([]int64, len(dbParts))
	for i, p := range dbParts {
		ids[i] = p.ID
	}
	return ids, nil
}

// BulkDeleteParts удаляет несколько запчастей с использованием worker pool для параллельной обработки
func (s *inventoryService) BulkDeleteParts(ctx context.Context, ids []int64) error {
	_, err := s.BulkDeletePartsWithProto(ctx, &partsv1.BulkDeletePartsRequest{
		Ids: ids,
	})
	return err
}

func collectPhotosToDelete(parts []domain.Part) []string {
	var photos []string
	for _, p := range parts {
		for _, photo := range p.Photos {
			if strings.TrimSpace(photo) != "" {
				photos = append(photos, photo)
			}
		}
	}
	return photos
}

func (s *inventoryService) resolveDeleteIDs(ctx context.Context, req *partsv1.BulkDeletePartsRequest) ([]int64, error) {
	if req.Scope != nil && req.Scope.Type == partsv1.BulkScopeType_BULK_SCOPE_TYPE_FILTER {
		if isFilterEmpty(req.Scope.Filter) && !req.ConfirmEntireCatalog {
			return nil, fmt.Errorf("массовое удаление всего каталога требует явного подтверждения (confirm_entire_catalog)")
		}
		resolved, err := s.resolvePartIDsFromFilter(ctx, req.Scope.Filter)
		if err != nil {
			return nil, fmt.Errorf("ошибка выборки запчастей для удаления: %w", err)
		}
		return resolved, nil
	}
	if req.Scope != nil && (req.Scope.Type == partsv1.BulkScopeType_BULK_SCOPE_TYPE_IDS || len(req.Scope.Ids) > 0) {
		return req.Scope.Ids, nil
	}
	return req.Ids, nil
}

// BulkDeletePartsWithProto выполняет массовое удаление запчастей по Protobuf спецификации
func (s *inventoryService) BulkDeletePartsWithProto(ctx context.Context, req *partsv1.BulkDeletePartsRequest) (int, error) {
	if req == nil {
		return 0, nil
	}

	ids, err := s.resolveDeleteIDs(ctx, req)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}

	slog.InfoContext(ctx, "InventoryService.BulkDeletePartsWithProto: Starting bulk delete", slog.Int("count", len(ids)))

	// 1. Считываем фотографии для последующего фонового удаления файлов с диска
	partsToDelete, err := s.repo.FindWithFilters(ctx, map[string]interface{}{"ids_in": ids}, 0, len(ids))
	var photosToDelete []string
	if err == nil {
		photosToDelete = collectPhotosToDelete(partsToDelete)
	}

	// 2. Мгновенное удаление записей в PostgreSQL (транзакция 50-150 мс)
	if err := s.repo.BulkDelete(ctx, ids); err != nil {
		slog.ErrorContext(ctx, "InventoryService.BulkDeletePartsWithProto: Failed to bulk delete in database", logger.Err(err))
		return 0, err
	}

	// 3. Пакетное удаление из Elasticsearch
	if err := search.BulkDeletePartsFromIndex(ctx, ids); err != nil {
		slog.WarnContext(ctx, "InventoryService.BulkDeletePartsWithProto: Failed to bulk delete from Elasticsearch", logger.Err(err))
	}

	// 4. Сброс Redis-кэша инвентаря и статистики
	if s.redis != nil {
		s.redis.Del(ctx, inventoryCacheKey)
		s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
		s.invalidateStatisticsCache(ctx)
	}

	// 5. Фоновая зачистка физических файлов фото с диска и отправка событий в Redis Streams
	if len(photosToDelete) > 0 || len(ids) > 0 {
		go func(photos []string, deletedIDs []int64) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			for _, photo := range photos {
				if err := DeletePhotoFile(photo); err != nil {
					slog.WarnContext(bgCtx, "Failed to delete photo file in background", logger.Err(err), slog.String("photo", photo))
				}
			}

			if s.redis != nil {
				for _, id := range deletedIDs {
					_ = s.redis.XAdd(bgCtx, &redis.XAddArgs{
						Stream: constants.StreamOrdersEvents,
						Values: map[string]interface{}{
							"type":    "part_delete_requested",
							"part_id": fmt.Sprintf("%d", id),
						},
					}).Err()
				}
			}
		}(photosToDelete, ids)
	}

	slog.InfoContext(ctx, "InventoryService.BulkDeletePartsWithProto: Successfully deleted parts", slog.Int("deleted_count", len(ids)))
	return len(ids), nil
}

func convertLegacyPartsUpdates(parts []*partsv1.BulkUpdatePartData) ([]map[string]interface{}, []int64) {
	updates := make([]map[string]interface{}, len(parts))
	ids := make([]int64, len(parts))
	for i, p := range parts {
		m := make(map[string]interface{}, len(p.Fields)+1)
		m["id"] = p.Id
		ids[i] = p.Id
		for k, v := range p.Fields {
			m[k] = v
		}
		updates[i] = m
	}
	return updates, ids
}

func (s *inventoryService) prepareBulkUpdateFilter(ctx context.Context, req *partsv1.BulkUpdatePartsRequest) (map[string]interface{}, bool, error) {
	if req.Scope.Type == partsv1.BulkScopeType_BULK_SCOPE_TYPE_FILTER {
		if isFilterEmpty(req.Scope.Filter) && !req.ConfirmEntireCatalog {
			return nil, false, fmt.Errorf("массовая операция над всем каталогом требует явного подтверждения (confirm_entire_catalog)")
		}

		params := inventoryQueryParamsFromProtoFilter(req.Scope.Filter)
		// Резолвим IDs через Elasticsearch ТОЛЬКО если задан свободный текстовый поиск (Search != "").
		// Для структурных фильтров склада (наличие фото, статус, категория, бренд, склад и т.д.)
		// запрос выполняется напрямую в PostgreSQL без искусственного лимита Elasticsearch в 10 000.
		if strings.TrimSpace(params.Search) != "" {
			resolvedIDs, err := s.resolvePartIDsFromFilter(ctx, req.Scope.Filter)
			if err != nil {
				return nil, false, fmt.Errorf("ошибка разрешения списка запчастей: %w", err)
			}
			if len(resolvedIDs) == 0 {
				return nil, true, nil
			}
			req.Scope.Type = partsv1.BulkScopeType_BULK_SCOPE_TYPE_IDS
			req.Scope.Ids = resolvedIDs
		}
	}

	var filterMap map[string]interface{}
	if req.Scope.Type == partsv1.BulkScopeType_BULK_SCOPE_TYPE_FILTER && req.Scope.Filter != nil {
		params := inventoryQueryParamsFromProtoFilter(req.Scope.Filter)
		filterMap = s.buildDatabaseFilters(params)
		filterMap["to_delete_at_is_null"] = true
		filterMap["quantity_gte"] = 0
	}
	return filterMap, false, nil
}

func (s *inventoryService) executeLegacyBulkUpdate(ctx context.Context, parts []*partsv1.BulkUpdatePartData) ([]domain.Part, int, error) {
	updates, ids := convertLegacyPartsUpdates(parts)
	count, err := s.repo.BulkUpdate(ctx, updates)
	if err != nil {
		slog.ErrorContext(ctx, "InventoryService.BulkUpdateParts: Failed legacy bulk update", logger.Err(err))
		return nil, 0, err
	}
	var updatedParts []domain.Part
	if len(ids) > 0 {
		if p, err := s.repo.FindWithFilters(ctx, map[string]interface{}{"ids_in": ids}, 0, 0); err == nil {
			updatedParts = p
		}
	}
	return updatedParts, count, nil
}

// BulkUpdateParts обновляет запчасти по Protobuf спецификации
func (s *inventoryService) BulkUpdateParts(ctx context.Context, req *partsv1.BulkUpdatePartsRequest) (int, error) {
	if req == nil {
		return 0, nil
	}

	slog.InfoContext(ctx, "InventoryService.BulkUpdateParts: Starting bulk update",
		slog.Bool("has_scope", req.Scope != nil),
		slog.Int("legacy_parts", len(req.Parts)))

	var updatedParts []domain.Part
	var updatedCount int

	// Сценарий 1: Умный Scope (по IDs или по Filter)
	if req.Scope != nil {
		filterMap, earlyExit, err := s.prepareBulkUpdateFilter(ctx, req)
		if err != nil {
			return 0, err
		}
		if earlyExit {
			return 0, nil
		}

		parts, err := s.repo.BulkUpdateWithProto(ctx, req, filterMap)
		if err != nil {
			slog.ErrorContext(ctx, "InventoryService.BulkUpdateParts: Failed to update with proto scope", logger.Err(err))
			return 0, err
		}
		updatedParts = parts
		updatedCount = len(parts)
	} else if len(req.Parts) > 0 {
		// Сценарий 2: Обратная совместимость с массивом parts: [{id, fields}]
		parts, count, err := s.executeLegacyBulkUpdate(ctx, req.Parts)
		if err != nil {
			return 0, err
		}
		updatedParts = parts
		updatedCount = count
	}

	// 1. Elasticsearch Bulk Indexing (синхронизация полнотекстового индекса)
	if len(updatedParts) > 0 {
		if err := search.BulkIndexParts(ctx, updatedParts); err != nil {
			slog.WarnContext(ctx, "InventoryService.BulkUpdateParts: Failed to sync parts with Elasticsearch", logger.Err(err))
		}
	}

	// 2. Сброс Redis-кэша инвентаря и статистики
	if s.redis != nil {
		s.redis.Del(ctx, inventoryCacheKey)
		s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
		s.invalidateStatisticsCache(ctx)
	}

	slog.InfoContext(ctx, "InventoryService.BulkUpdateParts: Successfully completed bulk update", slog.Int("updated_count", updatedCount))
	return updatedCount, nil
}
