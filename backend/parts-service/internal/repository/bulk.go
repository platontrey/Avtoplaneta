package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	partsv1 "avtoplaneta/gen/parts/v1"
	"avtoplaneta/pkg/logger"
	"parts-service/db/sqlc"
	"parts-service/internal/domain"
)

func (r *partRepository) bulkUpdateByIDs(ctx context.Context, req *partsv1.BulkUpdatePartsRequest) ([]domain.Part, error) {
	if len(req.Scope.Ids) == 0 {
		return []domain.Part{}, nil
	}
	ids := make([]int64, len(req.Scope.Ids))
	for i, id := range req.Scope.Ids {
		ids[i] = int64(id)
	}

	priceMode := "set"
	hasPrice := false
	priceVal := 0.0
	if req.Operations != nil && req.Operations.Price != nil {
		hasPrice = true
		priceVal = req.Operations.Price.Value
		switch req.Operations.Price.Mode {
		case partsv1.PriceOperationMode_PRICE_OPERATION_MODE_PERCENT:
			priceMode = "percent"
		case partsv1.PriceOperationMode_PRICE_OPERATION_MODE_OFFSET:
			priceMode = "offset"
		default:
			priceMode = "set"
		}
	}

	ops := req.GetOperations()
	sqlcParts, err := r.queries.BulkUpdateFieldsByIDs(ctx, sqlc.BulkUpdateFieldsByIDsParams{
		HasPrice:              hasPrice,
		PriceMode:             priceMode,
		PriceVal:              priceVal,
		HasName:               ops != nil && ops.Name != nil,
		Name:                  ops.GetName(),
		HasQuantity:           ops != nil && ops.Quantity != nil,
		Quantity:              ops.GetQuantity(),
		HasDescription:        ops != nil && ops.Description != nil,
		Description:           ops.GetDescription(),
		HasCategory:           ops != nil && ops.Category != nil,
		Category:              ops.GetCategory(),
		HasLocation:           ops != nil && ops.Location != nil,
		Location:              ops.GetLocation(),
		HasAddress:            ops != nil && ops.Address != nil,
		Address:               ops.GetAddress(),
		HasSalesman:           ops != nil && ops.Salesman != nil,
		Salesman:              ops.GetSalesman(),
		HasSellerID:           ops != nil && ops.SellerId != nil,
		SellerID:              ops.GetSellerId(),
		HasStatus:             ops != nil && ops.Status != nil,
		Status:                ops.GetStatus(),
		HasBrand:              ops != nil && ops.Brand != nil,
		Brand:                 ops.GetBrand(),
		HasModel:              ops != nil && ops.Model != nil,
		Model:                 ops.GetModel(),
		HasVin:                ops != nil && ops.Vin != nil,
		Vin:                   ops.GetVin(),
		HasBodyBrand:          ops != nil && ops.BodyBrand != nil,
		BodyBrand:             ops.GetBodyBrand(),
		HasEngineBrand:        ops != nil && ops.EngineBrand != nil,
		EngineBrand:           ops.GetEngineBrand(),
		HasCarReleaseDate:     ops != nil && ops.CarReleaseDate != nil,
		CarReleaseDate:        ops.GetCarReleaseDate(),
		HasCarReleasePeriod:   ops != nil && ops.CarReleasePeriod != nil,
		CarReleasePeriod:      ops.GetCarReleasePeriod(),
		HasFrontRear:          ops != nil && ops.FrontRear != nil,
		FrontRear:             ops.GetFrontRear(),
		HasLeftRight:          ops != nil && ops.LeftRight != nil,
		LeftRight:             ops.GetLeftRight(),
		HasTopBottom:          ops != nil && ops.TopBottom != nil,
		TopBottom:             ops.GetTopBottom(),
		HasNumber:             ops != nil && ops.Number != nil,
		Number:                ops.GetNumber(),
		HasManufacturer:       ops != nil && ops.Manufacturer != nil,
		Manufacturer:          ops.GetManufacturer(),
		HasManufacturerCode:   ops != nil && ops.ManufacturerCode != nil,
		ManufacturerCode:      ops.GetManufacturerCode(),
		HasOemCode:            ops != nil && ops.OemCode != nil,
		OemCode:               ops.GetOemCode(),
		HasColor:              ops != nil && ops.Color != nil,
		Color:                 ops.GetColor(),
		HasCondition:          ops != nil && ops.Condition != nil,
		Condition:             ops.GetCondition(),
		HasSupplierCode:       ops != nil && ops.SupplierCode != nil,
		SupplierCode:          ops.GetSupplierCode(),
		HasDefect:             ops != nil && ops.Defect != nil,
		Defect:                ops.GetDefect(),
		HasTransmission:       ops != nil && ops.Transmission != nil,
		Transmission:          ops.GetTransmission(),
		HasTransmissionModel:  ops != nil && ops.TransmissionModel != nil,
		TransmissionModel:     ops.GetTransmissionModel(),
		HasDrive:              ops != nil && ops.Drive != nil,
		Drive:                 ops.GetDrive(),
		HasWearPercentage:     ops != nil && ops.WearPercentage != nil,
		WearPercentage:        ops.GetWearPercentage(),
		HasSeason:             ops != nil && ops.Season != nil,
		Season:                ops.GetSeason(),
		HasDiameter:           ops != nil && ops.Diameter != nil,
		Diameter:              ops.GetDiameter(),
		HasWidth:              ops != nil && ops.Width != nil,
		Width:                 ops.GetWidth(),
		HasProfile:            ops != nil && ops.Profile != nil,
		Profile:               ops.GetProfile(),
		HasTireQuantity:       ops != nil && ops.TireQuantity != nil,
		TireQuantity:          ops.GetTireQuantity(),
		HasDrilling:           ops != nil && ops.Drilling != nil,
		Drilling:              ops.GetDrilling(),
		HasOffset:             ops != nil && ops.Offset != nil,
		OffsetVal:             ops.GetOffset(),
		HasCenterHoleDiameter: ops != nil && ops.CenterHoleDiameter != nil,
		CenterHoleDiameter:    ops.GetCenterHoleDiameter(),
		HasTireModel:          ops != nil && ops.TireModel != nil,
		TireModel:             ops.GetTireModel(),
		Ids:                   ids,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update by IDs: %w", err)
	}

	result := make([]domain.Part, len(sqlcParts))
	for i, p := range sqlcParts {
		result[i] = *sqlcPartToDomain(p)
	}
	return result, nil
}

func applyBulkOperations(builder squirrel.UpdateBuilder, ops *partsv1.BulkFieldOperations) squirrel.UpdateBuilder {
	if ops == nil {
		return builder
	}
	if ops.Price != nil {
		switch ops.Price.Mode {
		case partsv1.PriceOperationMode_PRICE_OPERATION_MODE_PERCENT:
			builder = builder.Set("price", squirrel.Expr("ROUND((price * (1.0 + ? / 100.0))::numeric, 2)::double precision", ops.Price.Value))
		case partsv1.PriceOperationMode_PRICE_OPERATION_MODE_OFFSET:
			builder = builder.Set("price", squirrel.Expr("GREATEST(0.0, price + ?)", ops.Price.Value))
		default:
			builder = builder.Set("price", ops.Price.Value)
		}
	}
	if ops.Name != nil {
		builder = builder.Set("name", ops.GetName())
	}
	if ops.Quantity != nil {
		builder = builder.Set("quantity", ops.GetQuantity())
	}
	if ops.Description != nil {
		builder = builder.Set("description", ops.GetDescription())
	}
	if ops.Category != nil {
		builder = builder.Set("category", ops.GetCategory())
	}
	if ops.Location != nil {
		builder = builder.Set("location", ops.GetLocation())
	}
	if ops.Address != nil {
		builder = builder.Set("address", ops.GetAddress())
	}
	if ops.Salesman != nil {
		builder = builder.Set("salesman", ops.GetSalesman())
	}
	if ops.SellerId != nil {
		builder = builder.Set("seller_id", ops.GetSellerId())
	}
	if ops.Status != nil {
		builder = builder.Set("status", ops.GetStatus())
	}
	if ops.Brand != nil {
		builder = builder.Set("brand", ops.GetBrand())
	}
	if ops.Model != nil {
		builder = builder.Set("model", ops.GetModel())
	}
	if ops.Vin != nil {
		builder = builder.Set("vin", ops.GetVin())
	}
	if ops.BodyBrand != nil {
		builder = builder.Set("body_brand", ops.GetBodyBrand())
	}
	if ops.EngineBrand != nil {
		builder = builder.Set("engine_brand", ops.GetEngineBrand())
	}
	if ops.CarReleaseDate != nil {
		builder = builder.Set("car_release_date", ops.GetCarReleaseDate())
	}
	if ops.CarReleasePeriod != nil {
		builder = builder.Set("car_release_period", ops.GetCarReleasePeriod())
	}
	if ops.FrontRear != nil {
		builder = builder.Set("front_rear", ops.GetFrontRear())
	}
	if ops.LeftRight != nil {
		builder = builder.Set("left_right", ops.GetLeftRight())
	}
	if ops.TopBottom != nil {
		builder = builder.Set("top_bottom", ops.GetTopBottom())
	}
	if ops.Number != nil {
		builder = builder.Set("number", ops.GetNumber())
	}
	if ops.Manufacturer != nil {
		builder = builder.Set("manufacturer", ops.GetManufacturer())
	}
	if ops.ManufacturerCode != nil {
		builder = builder.Set("manufacturer_code", ops.GetManufacturerCode())
	}
	if ops.OemCode != nil {
		builder = builder.Set("oem_code", ops.GetOemCode())
	}
	if ops.Color != nil {
		builder = builder.Set("color", ops.GetColor())
	}
	if ops.Condition != nil {
		builder = builder.Set("condition", ops.GetCondition())
	}
	if ops.SupplierCode != nil {
		builder = builder.Set("supplier_code", ops.GetSupplierCode())
	}
	if ops.Defect != nil {
		builder = builder.Set("defect", ops.GetDefect())
	}
	if ops.Transmission != nil {
		builder = builder.Set("transmission", ops.GetTransmission())
	}
	if ops.TransmissionModel != nil {
		builder = builder.Set("transmission_model", ops.GetTransmissionModel())
	}
	if ops.Drive != nil {
		builder = builder.Set("drive", ops.GetDrive())
	}
	if ops.WearPercentage != nil {
		builder = builder.Set("wear_percentage", ops.GetWearPercentage())
	}
	if ops.Season != nil {
		builder = builder.Set("season", ops.GetSeason())
	}
	if ops.Diameter != nil {
		builder = builder.Set("diameter", ops.GetDiameter())
	}
	if ops.Width != nil {
		builder = builder.Set("width", ops.GetWidth())
	}
	if ops.Profile != nil {
		builder = builder.Set("profile", ops.GetProfile())
	}
	if ops.TireQuantity != nil {
		builder = builder.Set("tire_quantity", ops.GetTireQuantity())
	}
	if ops.Drilling != nil {
		builder = builder.Set("drilling", ops.GetDrilling())
	}
	if ops.Offset != nil {
		builder = builder.Set(`"offset"`, ops.GetOffset())
	}
	if ops.CenterHoleDiameter != nil {
		builder = builder.Set("center_hole_diameter", ops.GetCenterHoleDiameter())
	}
	if ops.TireModel != nil {
		builder = builder.Set("tire_model", ops.GetTireModel())
	}
	return builder
}

func (r *partRepository) bulkUpdateByFilter(ctx context.Context, req *partsv1.BulkUpdatePartsRequest, filterMap map[string]interface{}) ([]domain.Part, error) {
	updateBuilder := r.psq.Update("parts").
		Where("deleted_at IS NULL").
		Where(squirrel.GtOrEq{"quantity": 0})

	if filterMap != nil {
		for _, pred := range r.buildFilterPredicates(filterMap) {
			updateBuilder = updateBuilder.Where(pred)
		}
	}

	updateBuilder = applyBulkOperations(updateBuilder, req.GetOperations())
	updateBuilder = updateBuilder.Set("updated_at", squirrel.Expr("NOW()"))
	updateBuilder = updateBuilder.Suffix("RETURNING " + partColumns)

	sqlStr, args, err := updateBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build bulk update query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute bulk update by filter: %w", err)
	}
	defer rows.Close()

	return scanParts(rows)
}

// BulkUpdateWithProto выполняет массовое обновление по спецификации Protobuf
func (r *partRepository) BulkUpdateWithProto(ctx context.Context, req *partsv1.BulkUpdatePartsRequest, filterMap map[string]interface{}) ([]domain.Part, error) {
	if req == nil {
		return nil, nil
	}

	// 1. Если указан scope по IDs (или просто непустой список IDs) - используем sqlc
	if req.Scope != nil && (req.Scope.Type == partsv1.BulkScopeType_BULK_SCOPE_TYPE_IDS || (req.Scope.Type == partsv1.BulkScopeType_BULK_SCOPE_TYPE_UNSPECIFIED && len(req.Scope.Ids) > 0)) {
		return r.bulkUpdateByIDs(ctx, req)
	}

	// 2. Если указан scope по фильтру (условиям)
	if req.Scope != nil && req.Scope.Type == partsv1.BulkScopeType_BULK_SCOPE_TYPE_FILTER {
		return r.bulkUpdateByFilter(ctx, req, filterMap)
	}

	return []domain.Part{}, nil
}

func (r *partRepository) BulkDelete(ctx context.Context, ids []int64) error {
	slog.InfoContext(ctx, "PartRepository.BulkDelete: Starting bulk delete",
		slog.Any("ids", ids),
		slog.Int("count", len(ids)))

	if err := r.queries.BulkDeleteByIDs(ctx, ids); err != nil {
		slog.ErrorContext(ctx, "PartRepository.BulkDelete: Failed to execute delete query", logger.Err(err))
		return err
	}

	slog.InfoContext(ctx, "PartRepository.BulkDelete: Successfully completed bulk delete")
	return nil
}

// BulkUpdate выполняет пакетное обновление произвольных полей запчастей через pgx.Batch и squirrel.
// Избегает оверхеда временных таблиц и исключает одиночные запросы в цикле (AGENTS.md п. 3.4.9).
func (r *partRepository) BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error) {
	slog.InfoContext(ctx, "PartRepository.BulkUpdate: Starting bulk update", slog.Int("count", len(updates)))

	if len(updates) == 0 {
		return 0, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin tx for bulk update: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	now := time.Now()

	for i, update := range updates {
		var id int64
		if idVal, exists := update["id"]; exists {
			switch v := idVal.(type) {
			case float64:
				id = int64(v)
			case int:
				id = int64(v)
			case int64:
				id = v
			case uint:
				id = int64(v)
			default:
				return 0, fmt.Errorf("update at index %d has invalid id type: %T", i, idVal)
			}
		} else {
			return 0, fmt.Errorf("update at index %d missing 'id' field", i)
		}

		builder := r.psq.Update("parts").Where(squirrel.Eq{"id": id, "deleted_at": nil})
		fieldCount := 0

		for key, value := range update {
			if key == "id" {
				continue
			}
			if key == "photos" {
				if arr, ok := value.(domain.StringArray); ok {
					jsonVal, _ := json.Marshal(arr)
					builder = builder.Set(key, jsonVal)
				} else {
					builder = builder.Set(key, value)
				}
			} else if key == "offset" {
				builder = builder.Set(`"offset"`, value)
			} else {
				builder = builder.Set(key, value)
			}
			fieldCount++
		}

		if fieldCount == 0 {
			continue
		}

		builder = builder.Set("updated_at", now)
		sqlStr, args, err := builder.ToSql()
		if err != nil {
			return 0, fmt.Errorf("failed to build update query at index %d: %w", i, err)
		}

		batch.Queue(sqlStr, args...)
	}

	if batch.Len() == 0 {
		return 0, nil
	}

	br := tx.SendBatch(ctx, batch)
	var updatedCount int

	for i := 0; i < batch.Len(); i++ {
		cmdTag, err := br.Exec()
		if err != nil {
			_ = br.Close()
			slog.ErrorContext(ctx, "PartRepository.BulkUpdate: Failed to execute batch item", logger.Err(err), slog.Int("index", i))
			return 0, fmt.Errorf("failed to execute batch update at index %d: %w", i, err)
		}
		updatedCount += int(cmdTag.RowsAffected())
	}

	if err := br.Close(); err != nil {
		return 0, fmt.Errorf("failed to close batch: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		slog.ErrorContext(ctx, "PartRepository.BulkUpdate: Failed to commit transaction", logger.Err(err))
		return 0, err
	}

	slog.InfoContext(ctx, "PartRepository.BulkUpdate: Successfully completed bulk update", slog.Int("updated_count", updatedCount))
	return updatedCount, nil
}
