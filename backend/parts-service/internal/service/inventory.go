package service

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"avtoplaneta/pkg/logger"
	"parts-service/internal/catalog"
	"parts-service/internal/domain"
	"parts-service/internal/search"
)

// GetInventory получает инвентарь запчастей с учетом фильтров и пагинации
func (s *inventoryService) GetInventory(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, error) {
	parts, _, err := s.GetInventoryWithCount(ctx, params)
	return parts, err
}

// GetInventoryWithCount получает инвентарь запчастей и общее количество с учетом фильтров
func (s *inventoryService) GetInventoryWithCount(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, int64, error) {
	// Очищаем просроченные запчасти перед поиском
	if err := s.cleanupExpiredParts(ctx); err != nil {
		// Логируем ошибку, но продолжаем выполнение
		slog.WarnContext(ctx, "Failed to cleanup expired parts during inventory fetch", logger.Err(err))
	}

	// Выбираем источник данных на основе параметров поиска
	parts, total, err := s.fetchPartsWithCount(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	// Форматируем дополнительные поля для фронтенда
	s.formatPartsForDisplay(parts)

	return parts, total, nil
}

// cleanupExpiredParts удаляет запчасти, отмеченные для удаления более 14 дней назад
func (s *inventoryService) cleanupExpiredParts(ctx context.Context) error {
	fourteenDaysAgo := time.Now().AddDate(0, 0, -14)
	err := s.repo.DeleteExpiredParts(ctx, fourteenDaysAgo)
	if err != nil {
		slog.WarnContext(ctx, "Failed to cleanup expired parts", logger.Err(err))
	}
	return err
}

// fetchParts выбирает оптимальный источник данных для поиска
func (s *inventoryService) fetchParts(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, error) {
	parts, _, err := s.fetchPartsWithCount(ctx, params)
	return parts, err
}

// fetchPartsWithCount выбирает оптимальный источник данных для поиска и возвращает общее количество
func (s *inventoryService) fetchPartsWithCount(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, int64, error) {
	if s.shouldUseElasticsearch(params) && s.es != nil {
		return s.getInventoryFromElasticsearchWithCount(ctx, params)
	}
	return s.getInventoryFromDatabaseWithCount(ctx, params)
}

// formatPartsForDisplay добавляет форматированные поля для отображения
func (s *inventoryService) formatPartsForDisplay(parts []domain.Part) {
	now := time.Now()
	for i := range parts {
		if parts[i].ToDeleteAt != nil {
			parts[i].ToDeleteAtFormatted = parts[i].ToDeleteAt.Format("2006-01-02 15:04:05")
			parts[i].TimeUntilDeletion = domain.FormatTimeUntil(*parts[i].ToDeleteAt, now)
		}
	}
}

// shouldUseElasticsearch определяет, использовать ли Elasticsearch
func (s *inventoryService) shouldUseElasticsearch(params domain.InventoryQueryParams) bool {
	return params.Search != "" || params.Category != "" || params.Brand != "" ||
		params.Model != "" || params.Location != "" || params.Address != "" || params.Salesman != "" ||
		params.Status != "" || params.HasPhoto != "" || params.Number != "" || params.OEMCode != "" ||
		params.VIN != "" || params.BodyBrand != "" || params.EngineBrand != "" ||
		params.CarReleaseDate != "" || params.CarReleasePeriod != "" || params.Transmission != "" || params.Drive != "" ||
		params.Condition != "" || params.Manufacturer != "" || params.Defect != "" || params.Color != "" ||
		params.MinPrice != "" || params.MaxPrice != "" || params.MinQuantity != "" || params.MaxQuantity != "" ||
		params.FrontRear != "" || params.LeftRight != "" || params.TopBottom != "" ||
		params.ManufacturerCode != "" || params.SupplierCode != "" || params.TransmissionModel != "" ||
		params.WearPercentage != "" || params.Season != "" || params.Diameter != "" ||
		params.Width != "" || params.Profile != "" || params.TireQuantity != "" ||
		params.Drilling != "" || params.Offset != "" || params.CenterHoleDiameter != "" || params.TireModel != ""
}

// getInventoryFromElasticsearch получает данные из Elasticsearch
func (s *inventoryService) getInventoryFromElasticsearch(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, error) {
	parts, _, err := s.getInventoryFromElasticsearchWithCount(ctx, params)
	return parts, err
}

// getInventoryFromElasticsearchWithCount получает данные из Elasticsearch и общее количество
func (s *inventoryService) getInventoryFromElasticsearchWithCount(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, int64, error) {
	esQuery := s.buildElasticsearchQuery(params)
	from := (params.Page - 1) * params.Limit

	esParts, totalES, err := s.es.SearchParts(esQuery, from, params.Limit)
	if err != nil {
		slog.WarnContext(ctx, "Error searching with Elasticsearch, falling back to database", logger.Err(err))
		// Fallback to database
		return s.getInventoryFromDatabaseWithCount(ctx, params)
	}

	// Преобразуем и фильтруем
	if len(esParts) == 0 {
		// Fallback to database to ensure we don't return 0 if DB actually has matching parts
		dbParts, dbCount, dbErr := s.getInventoryFromDatabaseWithCount(ctx, params)
		if dbErr == nil && len(dbParts) > 0 {
			return dbParts, dbCount, nil
		}
		return []domain.Part{}, totalES, nil
	}

	parts := make([]domain.Part, 0, len(esParts))
	partIDs := make([]int64, len(esParts))
	for i, esPart := range esParts {
		partIDs[i] = esPart.ID
	}

	// Получаем валидные части из БД
	filters := map[string]interface{}{
		"ids_in":               partIDs,
		"to_delete_at_is_null": true,
		"quantity_gte":         0,
	}
	validParts, err := s.repo.FindWithFilters(ctx, filters, 0, 0)
	if err != nil {
		return nil, 0, err
	}

	validPartsMap := make(map[int64]domain.Part, len(validParts))
	for _, part := range validParts {
		validPartsMap[part.ID] = part
	}

	parts = make([]domain.Part, 0, len(esParts))
	for _, esPart := range esParts {
		if part, ok := validPartsMap[esPart.ID]; ok {
			parts = append(parts, part)
		}
	}

	return parts, totalES, nil
}

// buildElasticsearchQuery строит запрос для Elasticsearch (делегирует в search.BuildElasticsearchQuery)
func (s *inventoryService) buildElasticsearchQuery(params domain.InventoryQueryParams) map[string]interface{} {
	return search.BuildElasticsearchQuery(params)
}

// getInventoryFromDatabase получает данные из базы данных
func (s *inventoryService) getInventoryFromDatabase(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, error) {
	parts, _, err := s.getInventoryFromDatabaseWithCount(ctx, params)
	return parts, err
}

// getInventoryFromDatabaseWithCount получает данные из базы данных и общее количество
func (s *inventoryService) getInventoryFromDatabaseWithCount(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, int64, error) {
	filters := s.buildDatabaseFilters(params)
	// Add default filters
	filters["to_delete_at_is_null"] = true
	filters["quantity_gte"] = 0

	// Calculate offset for pagination
	offset := 0
	limit := 0
	if params.Limit > 0 {
		offset = (params.Page - 1) * params.Limit
		limit = params.Limit
	}

	parts, err := s.repo.FindWithFilters(ctx, filters, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.repo.CountWithFilters(ctx, filters)
	if err != nil {
		slog.WarnContext(ctx, "Failed to count parts with filters, falling back to len(parts)", logger.Err(err))
		count = int64(len(parts))
	}

	return parts, count, nil
}

// buildDatabaseFilters строит фильтры для базы данных
func (s *inventoryService) buildDatabaseFilters(params domain.InventoryQueryParams) map[string]interface{} {
	filters := make(map[string]interface{})

	if params.Search != "" {
		filters["search"] = params.Search
	}
	if params.Category != "" {
		filters["category_ilike"] = catalog.ExpandCategorySynonyms(params.Category)
	}
	if params.Brand != "" {
		filters["brand_ilike"] = params.Brand
	}
	if params.Model != "" {
		filters["model_ilike"] = params.Model
	}
	if params.Location != "" {
		filters["location_ilike"] = params.Location
	}
	if params.Address != "" {
		filters["address_ilike"] = params.Address
	}
	if params.Salesman != "" {
		filters["salesman_ilike"] = params.Salesman
	}
	if params.Status != "" {
		filters["status"] = params.Status
	}
	if params.HasPhoto != "" && params.HasPhoto != "all" {
		var hasPhotoBool bool
		if params.HasPhoto == "with" {
			hasPhotoBool = true
		} else if params.HasPhoto == "without" {
			hasPhotoBool = false
		}
		filters["has_photo"] = hasPhotoBool
	}
	if params.Number != "" {
		filters["number_ilike"] = params.Number
	}
	if params.OEMCode != "" {
		filters["oem_code_ilike"] = params.OEMCode
	}
	if params.VIN != "" {
		filters["vin_ilike"] = params.VIN
	}
	if params.BodyBrand != "" {
		filters["body_brand_ilike"] = params.BodyBrand
	}
	if params.EngineBrand != "" {
		filters["engine_brand_ilike"] = params.EngineBrand
	}
	if params.CarReleaseDate != "" {
		filters["car_release_date_ilike"] = params.CarReleaseDate
	}
	if params.CarReleasePeriod != "" {
		filters["car_release_period_ilike"] = params.CarReleasePeriod
	}
	if params.Transmission != "" {
		filters["transmission_ilike"] = params.Transmission
	}
	if params.Drive != "" {
		filters["drive_ilike"] = params.Drive
	}
	if params.Condition != "" {
		filters["condition_ilike"] = params.Condition
	}
	if params.Manufacturer != "" {
		filters["manufacturer_ilike"] = params.Manufacturer
	}
	if params.Defect != "" {
		filters["defect_ilike"] = params.Defect
	}
	if params.Color != "" {
		filters["color_ilike"] = params.Color
	}
	if params.MinPrice != "" {
		if val, err := strconv.ParseFloat(params.MinPrice, 64); err == nil {
			filters["min_price"] = val
		}
	}
	if params.MaxPrice != "" {
		if val, err := strconv.ParseFloat(params.MaxPrice, 64); err == nil {
			filters["max_price"] = val
		}
	}
	if params.MinQuantity != "" {
		if val, err := strconv.Atoi(params.MinQuantity); err == nil {
			filters["min_quantity"] = val
		}
	}
	if params.MaxQuantity != "" {
		if val, err := strconv.Atoi(params.MaxQuantity); err == nil {
			filters["max_quantity"] = val
		}
	}
	if params.FrontRear != "" {
		filters["front_rear_ilike"] = params.FrontRear
	}
	if params.LeftRight != "" {
		filters["left_right_ilike"] = params.LeftRight
	}
	if params.TopBottom != "" {
		filters["top_bottom_ilike"] = params.TopBottom
	}
	if params.ManufacturerCode != "" {
		filters["manufacturer_code_ilike"] = params.ManufacturerCode
	}
	if params.SupplierCode != "" {
		filters["supplier_code_ilike"] = params.SupplierCode
	}
	if params.TransmissionModel != "" {
		filters["transmission_model_ilike"] = params.TransmissionModel
	}
	if params.WearPercentage != "" {
		filters["wear_percentage_ilike"] = params.WearPercentage
	}
	if params.Season != "" {
		filters["season_ilike"] = params.Season
	}
	if params.Diameter != "" {
		filters["diameter_ilike"] = params.Diameter
	}
	if params.Width != "" {
		filters["width_ilike"] = params.Width
	}
	if params.Profile != "" {
		filters["profile_ilike"] = params.Profile
	}
	if params.TireQuantity != "" {
		filters["tire_quantity_ilike"] = params.TireQuantity
	}
	if params.Drilling != "" {
		filters["drilling_ilike"] = params.Drilling
	}
	if params.Offset != "" {
		filters["offset_ilike"] = params.Offset
	}
	if params.CenterHoleDiameter != "" {
		filters["center_hole_diameter_ilike"] = params.CenterHoleDiameter
	}
	if params.TireModel != "" {
		filters["tire_model_ilike"] = params.TireModel
	}

	return filters
}
