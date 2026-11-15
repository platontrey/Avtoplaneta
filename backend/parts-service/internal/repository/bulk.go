package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Masterminds/squirrel"

	"avtoplaneta/pkg/logger"
	partsv1 "avtoplaneta/gen/parts/v1"
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

func (r *partRepository) BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error) {
	slog.InfoContext(ctx, "PartRepository.BulkUpdate: Starting bulk update", slog.Int("count", len(updates)))

	if len(updates) == 0 {
		return 0, nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Создаем временную таблицу
	_, err = tx.Exec(ctx, `
		CREATE TEMP TABLE temp_parts_update (
			id BIGINT PRIMARY KEY,
			name TEXT, quantity INTEGER, description TEXT, category TEXT,
			price DOUBLE PRECISION, brand TEXT, model TEXT, location TEXT, address TEXT,
			salesman TEXT, status TEXT, photos JSONB,
			body_brand TEXT, engine_brand TEXT, car_release_date TEXT, car_release_period TEXT,
			front_rear TEXT, left_right TEXT, top_bottom TEXT,
			number TEXT, manufacturer TEXT, manufacturer_code TEXT,
			oem_code TEXT, color TEXT, condition TEXT,
			supplier_code TEXT, defect TEXT, transmission TEXT, transmission_model TEXT,
			drive TEXT, wear_percentage TEXT,
			season TEXT, diameter TEXT, width TEXT, profile TEXT,
			tire_quantity TEXT, drilling TEXT, "offset" TEXT,
			center_hole_diameter TEXT, tire_model TEXT
		) ON COMMIT DROP
	`)
	if err != nil {
		slog.ErrorContext(ctx, "PartRepository.BulkUpdate: Failed to create temp table", logger.Err(err))
		return 0, err
	}

	// Вставляем данные во временную таблицу
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

		delete(update, "id")

		_, err = tx.Exec(ctx, `
			INSERT INTO temp_parts_update (id, name, quantity, description, category, price, brand, model, location, address, salesman, status, photos,
				body_brand, engine_brand, car_release_date, car_release_period, front_rear, left_right, top_bottom, number, manufacturer,
				manufacturer_code, oem_code, color, condition, supplier_code, defect, transmission, transmission_model, drive, wear_percentage,
				season, diameter, width, profile, tire_quantity, drilling, "offset", center_hole_diameter, tire_model)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37,$38,$39,$40,$41)`,
			id,
			update["name"], update["quantity"], update["description"], update["category"],
			update["price"], update["brand"], update["model"], update["location"], update["address"],
			update["salesman"], update["status"], update["photos"],
			update["body_brand"], update["engine_brand"], update["car_release_date"], update["car_release_period"],
			update["front_rear"], update["left_right"], update["top_bottom"],
			update["number"], update["manufacturer"], update["manufacturer_code"],
			update["oem_code"], update["color"], update["condition"],
			update["supplier_code"], update["defect"], update["transmission"], update["transmission_model"],
			update["drive"], update["wear_percentage"],
			update["season"], update["diameter"], update["width"], update["profile"],
			update["tire_quantity"], update["drilling"], update["offset"],
			update["center_hole_diameter"], update["tire_model"],
		)
		if err != nil {
			slog.ErrorContext(ctx, "PartRepository.BulkUpdate: Failed to insert into temp table", logger.Err(err), slog.Int("index", i))
			return 0, err
		}
	}

	// Batch UPDATE из temp таблицы
	_, err = tx.Exec(ctx, `
		UPDATE parts SET
			name = COALESCE(t.name, parts.name),
			quantity = COALESCE(t.quantity, parts.quantity),
			description = COALESCE(t.description, parts.description),
			category = COALESCE(t.category, parts.category),
			price = COALESCE(t.price, parts.price),
			brand = COALESCE(t.brand, parts.brand),
			model = COALESCE(t.model, parts.model),
			location = COALESCE(t.location, parts.location),
			address = COALESCE(t.address, parts.address),
			salesman = COALESCE(t.salesman, parts.salesman),
			status = COALESCE(t.status, parts.status),
			photos = COALESCE(t.photos, parts.photos),
			body_brand = COALESCE(t.body_brand, parts.body_brand),
			engine_brand = COALESCE(t.engine_brand, parts.engine_brand),
			car_release_date = COALESCE(t.car_release_date, parts.car_release_date),
			car_release_period = COALESCE(t.car_release_period, parts.car_release_period),
			front_rear = COALESCE(t.front_rear, parts.front_rear),
			left_right = COALESCE(t.left_right, parts.left_right),
			top_bottom = COALESCE(t.top_bottom, parts.top_bottom),
			number = COALESCE(t.number, parts.number),
			manufacturer = COALESCE(t.manufacturer, parts.manufacturer),
			manufacturer_code = COALESCE(t.manufacturer_code, parts.manufacturer_code),
			oem_code = COALESCE(t.oem_code, parts.oem_code),
			color = COALESCE(t.color, parts.color),
			condition = COALESCE(t.condition, parts.condition),
			supplier_code = COALESCE(t.supplier_code, parts.supplier_code),
			defect = COALESCE(t.defect, parts.defect),
			transmission = COALESCE(t.transmission, parts.transmission),
			transmission_model = COALESCE(t.transmission_model, parts.transmission_model),
			drive = COALESCE(t.drive, parts.drive),
			wear_percentage = COALESCE(t.wear_percentage, parts.wear_percentage),
			season = COALESCE(t.season, parts.season),
			diameter = COALESCE(t.diameter, parts.diameter),
			width = COALESCE(t.width, parts.width),
			profile = COALESCE(t.profile, parts.profile),
			tire_quantity = COALESCE(t.tire_quantity, parts.tire_quantity),
			drilling = COALESCE(t.drilling, parts.drilling),
			"offset" = COALESCE(t."offset", parts."offset"),
			center_hole_diameter = COALESCE(t.center_hole_diameter, parts.center_hole_diameter),
			tire_model = COALESCE(t.tire_model, parts.tire_model),
			updated_at = NOW()
		FROM temp_parts_update t
		WHERE parts.id = t.id
	`)
	if err != nil {
		slog.ErrorContext(ctx, "PartRepository.BulkUpdate: Failed to execute batch update", logger.Err(err))
		return 0, err
	}

	// Количество обновленных строк
	var updatedCount int64
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM temp_parts_update").Scan(&updatedCount)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		slog.ErrorContext(ctx, "PartRepository.BulkUpdate: Failed to commit transaction", logger.Err(err))
		return 0, err
	}

	slog.InfoContext(ctx, "PartRepository.BulkUpdate: Successfully completed bulk update", slog.Int64("updated_count", updatedCount))
	return int(updatedCount), nil
}
