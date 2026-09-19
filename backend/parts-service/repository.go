package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"parts-service/db/sqlc"
)

// PartRepository определяет контракт для доступа к данным запчастей
type PartRepository interface {
	Create(ctx context.Context, part *Part) error
	CreateBatch(ctx context.Context, parts []Part) ([]Part, error)
	FindByID(ctx context.Context, id int64) (*Part, error)
	FindAll(ctx context.Context) ([]Part, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	Delete(ctx context.Context, id int64) error

	DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error
	IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error

	FindWithFilters(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]Part, error)

	MarkForDeletion(ctx context.Context, id int64, deleteAt time.Time) error
	DeleteExpiredParts(ctx context.Context, before time.Time) error
	GetStatistics(ctx context.Context) (StatisticsResponse, error)
	InventoryVersion(ctx context.Context) (string, error)
	RenameSeller(ctx context.Context, sellerID int64, name string) ([]Part, error)
	BulkDelete(ctx context.Context, ids []int64) error
	BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error)

	DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error)
	GetSupplierCodes(ctx context.Context) ([]string, error)

	GetTotalEarnings(ctx context.Context) (float64, error)
	UpdateTotalEarnings(ctx context.Context, amount float64) error

	UpdatePartPhotos(ctx context.Context, id int64, photos StringArray) error
	GetPartsForXML(ctx context.Context) ([]Part, error)
	GetLastCreatedPart(ctx context.Context) (*Part, error)
}

// partRepository реализует PartRepository
type partRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
	psq     squirrel.StatementBuilderType
}

// NewPartRepository создает новый экземпляр репозитория
func NewPartRepository(pool *pgxpool.Pool) PartRepository {
	return &partRepository{
		pool:    pool,
		queries: sqlc.New(pool),
		psq:     squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// ─── Helpers ────────────────────────────────────────────────────────────────

var partColumns = `id, name, quantity, description, category, price, salesman, location, address, status,
	brand, model, photos, seller_id, to_delete_at, vin,
	body_brand, engine_brand, car_release_date, car_release_period, front_rear, left_right, top_bottom,
	number, manufacturer, manufacturer_code, oem_code, color, condition,
	supplier_code, defect, transmission, transmission_model, drive, wear_percentage,
	season, diameter, width, profile, tire_quantity, drilling, "offset",
	center_hole_diameter, tire_model, created_at, updated_at, deleted_at`

func scanPart(row pgx.Row) (*Part, error) {
	var p Part
	var photos []byte
	var toDeleteAt *time.Time
	var createdAt, updatedAt time.Time
	var deletedAt *time.Time

	err := row.Scan(
		&p.ID, &p.Name, &p.Quantity, &p.Description, &p.Category, &p.Price,
		&p.Salesman, &p.Location, &p.Address, &p.Status, &p.Brand, &p.Model, &photos,
		&p.SellerID, &toDeleteAt, &p.VIN,
		&p.BodyBrand, &p.EngineBrand, &p.CarReleaseDate, &p.CarReleasePeriod, &p.FrontRear, &p.LeftRight, &p.TopBottom,
		&p.Number, &p.Manufacturer, &p.ManufacturerCode, &p.OEMCode, &p.Color, &p.Condition,
		&p.SupplierCode, &p.Defect, &p.Transmission, &p.TransmissionModel, &p.Drive, &p.WearPercentage,
		&p.Season, &p.Diameter, &p.Width, &p.Profile, &p.TireQuantity, &p.Drilling, &p.Offset,
		&p.CenterHoleDiameter, &p.TireModel,
		&createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}

	p.ToDeleteAt = toDeleteAt
	if photos != nil {
		_ = json.Unmarshal(photos, &p.Photos)
	}
	if len(p.Photos) > 0 {
		p.Photo = p.Photos[0]
	}
	return &p, nil
}

func scanParts(rows pgx.Rows) ([]Part, error) {
	var parts []Part
	for rows.Next() {
		var p Part
		var photos []byte
		var toDeleteAt *time.Time
		var createdAt, updatedAt time.Time
		var deletedAt *time.Time

		err := rows.Scan(
			&p.ID, &p.Name, &p.Quantity, &p.Description, &p.Category, &p.Price,
			&p.Salesman, &p.Location, &p.Address, &p.Status, &p.Brand, &p.Model, &photos,
			&p.SellerID, &toDeleteAt, &p.VIN,
			&p.BodyBrand, &p.EngineBrand, &p.CarReleaseDate, &p.CarReleasePeriod, &p.FrontRear, &p.LeftRight, &p.TopBottom,
			&p.Number, &p.Manufacturer, &p.ManufacturerCode, &p.OEMCode, &p.Color, &p.Condition,
			&p.SupplierCode, &p.Defect, &p.Transmission, &p.TransmissionModel, &p.Drive, &p.WearPercentage,
			&p.Season, &p.Diameter, &p.Width, &p.Profile, &p.TireQuantity, &p.Drilling, &p.Offset,
			&p.CenterHoleDiameter, &p.TireModel,
			&createdAt, &updatedAt, &deletedAt,
		)
		if err != nil {
			return nil, err
		}

		p.ToDeleteAt = toDeleteAt
		if photos != nil {
			_ = json.Unmarshal(photos, &p.Photos)
		}
		if len(p.Photos) > 0 {
			p.Photo = p.Photos[0]
		}
		parts = append(parts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if parts == nil {
		parts = []Part{}
	}
	return parts, nil
}

func sqlcPartToDomain(p sqlc.Part) *Part {
	res := &Part{
		PartCore: PartCore{
			ID:          p.ID,
			Name:        p.Name,
			Quantity:    int(p.Quantity),
			Description: p.Description,
			Category:    p.Category,
			Price:       p.Price,
			Salesman:    p.Salesman,
			Location:    p.Location,
			Address:     p.Address,
			Status:      p.Status,
			Brand:       p.Brand,
			Model:       p.Model,
			SellerID:    p.SellerID,
			VIN:         p.Vin,
		},
		PartSpecifications: PartSpecifications{
			BodyBrand:         p.BodyBrand,
			EngineBrand:       p.EngineBrand,
			CarReleaseDate:    p.CarReleaseDate,
			CarReleasePeriod:  p.CarReleasePeriod,
			FrontRear:         p.FrontRear,
			LeftRight:         p.LeftRight,
			TopBottom:         p.TopBottom,
			Number:            p.Number,
			Manufacturer:      p.Manufacturer,
			ManufacturerCode:  p.ManufacturerCode,
			OEMCode:           p.OemCode,
			Color:             p.Color,
			Condition:        p.Condition,
			SupplierCode:      p.SupplierCode,
			Defect:            p.Defect,
			Transmission:      p.Transmission,
			TransmissionModel: p.TransmissionModel,
			Drive:             p.Drive,
			WearPercentage:    p.WearPercentage,
		},
		PartTireSpecifications: PartTireSpecifications{
			Season:             p.Season,
			Diameter:           p.Diameter,
			Width:              p.Width,
			Profile:            p.Profile,
			TireQuantity:       p.TireQuantity,
			Drilling:           p.Drilling,
			Offset:             p.Offset,
			CenterHoleDiameter: p.CenterHoleDiameter,
			TireModel:          p.TireModel,
		},
	}
	if p.ToDeleteAt.Valid {
		t := p.ToDeleteAt.Time
		res.ToDeleteAt = &t
	}
	if p.Photos != nil {
		_ = json.Unmarshal(p.Photos, &res.Photos)
	}
	if len(res.Photos) > 0 {
		res.Photo = res.Photos[0]
	}
	return res
}

// ─── CRUD ───────────────────────────────────────────────────────────────────

func (r *partRepository) Create(ctx context.Context, part *Part) error {
	photosJSON, _ := json.Marshal(part.Photos)
	if part.Photo != "" && len(part.Photos) == 0 {
		photosJSON, _ = json.Marshal(StringArray{part.Photo})
	}

	var toDeleteAt pgtype.Timestamptz
	if part.ToDeleteAt != nil {
		toDeleteAt = pgtype.Timestamptz{Time: *part.ToDeleteAt, Valid: true}
	}

	created, err := r.queries.CreatePart(ctx, sqlc.CreatePartParams{
		Name:               part.Name,
		Quantity:           int32(part.Quantity),
		Description:        part.Description,
		Category:           part.Category,
		Price:              part.Price,
		Salesman:           part.Salesman,
		Location:           part.Location,
		Address:            part.Address,
		Status:             part.Status,
		Brand:              part.Brand,
		Model:              part.Model,
		Photos:             photosJSON,
		SellerID:           part.SellerID,
		ToDeleteAt:         toDeleteAt,
		Vin:                part.VIN,
		BodyBrand:          part.BodyBrand,
		EngineBrand:        part.EngineBrand,
		CarReleaseDate:     part.CarReleaseDate,
		CarReleasePeriod:   part.CarReleasePeriod,
		FrontRear:          part.FrontRear,
		LeftRight:          part.LeftRight,
		TopBottom:          part.TopBottom,
		Number:             part.Number,
		Manufacturer:       part.Manufacturer,
		ManufacturerCode:   part.ManufacturerCode,
		OemCode:            part.OEMCode,
		Color:              part.Color,
		Condition:          part.Condition,
		SupplierCode:       part.SupplierCode,
		Defect:             part.Defect,
		Transmission:       part.Transmission,
		TransmissionModel:  part.TransmissionModel,
		Drive:              part.Drive,
		WearPercentage:     part.WearPercentage,
		Season:             part.Season,
		Diameter:           part.Diameter,
		Width:              part.Width,
		Profile:            part.Profile,
		TireQuantity:       part.TireQuantity,
		Drilling:           part.Drilling,
		Offset:             part.Offset,
		CenterHoleDiameter: part.CenterHoleDiameter,
		TireModel:          part.TireModel,
	})
	if err != nil {
		return err
	}
	part.ID = created.ID
	return nil
}

func (r *partRepository) FindByID(ctx context.Context, id int64) (*Part, error) {
	p, err := r.queries.GetPartByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return sqlcPartToDomain(p), nil
}

func (r *partRepository) FindAll(ctx context.Context) ([]Part, error) {
	parts, err := r.queries.GetAllParts(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]Part, len(parts))
	for i, p := range parts {
		res[i] = *sqlcPartToDomain(p)
	}
	return res, nil
}

func (r *partRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	builder := r.psq.Update("parts").Where(squirrel.Eq{"id": id, "deleted_at": nil})
	for key, value := range updates {
		if key == "photos" {
			if arr, ok := value.(StringArray); ok {
				jsonVal, _ := json.Marshal(arr)
				builder = builder.Set(key, jsonVal)
			} else {
				builder = builder.Set(key, value)
			}
		} else if key == "offset" {
			builder = builder.Set("\"offset\"", value)
		} else {
			builder = builder.Set(key, value)
		}
	}
	builder = builder.Set("updated_at", time.Now())

	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *partRepository) CreateBatch(ctx context.Context, parts []Part) ([]Part, error) {
	if len(parts) == 0 {
		return parts, nil
	}

	batch := &pgx.Batch{}
	for i := range parts {
		part := &parts[i]
		photosJSON, _ := json.Marshal(part.Photos)
		if part.Photo != "" && len(part.Photos) == 0 {
			photosJSON, _ = json.Marshal(StringArray{part.Photo})
		}
		var toDeleteAt *time.Time
		if part.ToDeleteAt != nil {
			toDeleteAt = part.ToDeleteAt
		}

		batch.Queue(sqlc.CreatePart,
			part.Name, part.Quantity, part.Description, part.Category, part.Price,
			part.Salesman, part.Location, part.Address, part.Status, part.Brand, part.Model,
			photosJSON, part.SellerID, toDeleteAt, part.VIN,
			part.BodyBrand, part.EngineBrand, part.CarReleaseDate, part.CarReleasePeriod, part.FrontRear, part.LeftRight, part.TopBottom,
			part.Number, part.Manufacturer, part.ManufacturerCode, part.OEMCode, part.Color, part.Condition,
			part.SupplierCode, part.Defect, part.Transmission, part.TransmissionModel, part.Drive, part.WearPercentage,
			part.Season, part.Diameter, part.Width, part.Profile, part.TireQuantity, part.Drilling, part.Offset,
			part.CenterHoleDiameter, part.TireModel,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	var createdAt, updatedAt time.Time
	for i := range parts {
		err := br.QueryRow().Scan(
			&parts[i].ID, &parts[i].Name, &parts[i].Quantity, &parts[i].Description, &parts[i].Category,
			&parts[i].Price, &parts[i].Salesman, &parts[i].Location, &parts[i].Address, &parts[i].Status,
			&parts[i].Brand, &parts[i].Model, new([]byte), &parts[i].SellerID, new(*time.Time), &parts[i].VIN,
			&parts[i].BodyBrand, &parts[i].EngineBrand, &parts[i].CarReleaseDate, &parts[i].FrontRear,
			&parts[i].LeftRight, &parts[i].TopBottom, &parts[i].Number, &parts[i].Manufacturer,
			&parts[i].ManufacturerCode, &parts[i].OEMCode, &parts[i].Color, &parts[i].Condition,
			&parts[i].SupplierCode, &parts[i].Defect, &parts[i].Transmission, &parts[i].TransmissionModel,
			&parts[i].Drive, &parts[i].WearPercentage, &parts[i].Season, &parts[i].Diameter, &parts[i].Width,
			&parts[i].Profile, &parts[i].TireQuantity, &parts[i].Drilling, &parts[i].Offset,
			&parts[i].CenterHoleDiameter, &parts[i].TireModel, &createdAt, &updatedAt, new(*time.Time),
			&parts[i].CarReleasePeriod,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan created part %d in batch: %w", i, err)
		}
	}

	return parts, nil
}

func (r *partRepository) DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	if operationID != "" {
		tx, err := r.pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		q := r.queries.WithTx(tx)
		rowsAffected, err := q.RecordStockOperation(ctx, sqlc.RecordStockOperationParams{
			OperationID:   operationID,
			PartID:        id,
			OperationType: "decrease",
			Amount:        int32(amount),
		})
		if err != nil {
			return fmt.Errorf("failed to record stock operation: %w", err)
		}

		if rowsAffected == 0 {
			logrus.WithFields(logrus.Fields{
				"operation_id": operationID,
				"part_id":      id,
			}).Info("Stock operation already recorded, skipping decrease (idempotent)")
			return nil
		}

		if err := q.DecreasePartQuantity(ctx, sqlc.DecreasePartQuantityParams{
			ID:     id,
			Amount: int32(amount),
		}); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}

	return r.queries.DecreasePartQuantity(ctx, sqlc.DecreasePartQuantityParams{
		ID:     id,
		Amount: int32(amount),
	})
}

func (r *partRepository) IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	if operationID != "" {
		tx, err := r.pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		q := r.queries.WithTx(tx)
		rowsAffected, err := q.RecordStockOperation(ctx, sqlc.RecordStockOperationParams{
			OperationID:   operationID,
			PartID:        id,
			OperationType: "increase",
			Amount:        int32(amount),
		})
		if err != nil {
			return fmt.Errorf("failed to record stock operation: %w", err)
		}

		if rowsAffected == 0 {
			logrus.WithFields(logrus.Fields{
				"operation_id": operationID,
				"part_id":      id,
			}).Info("Stock operation already recorded, skipping increase (idempotent)")
			return nil
		}

		if err := q.IncreasePartQuantity(ctx, sqlc.IncreasePartQuantityParams{
			ID:     id,
			Amount: int32(amount),
		}); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}

	return r.queries.IncreasePartQuantity(ctx, sqlc.IncreasePartQuantityParams{
		ID:     id,
		Amount: int32(amount),
	})
}

func (r *partRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeletePart(ctx, id)
}

// ─── FindWithFilters (Squirrel dynamic query) ───────────────────────────────

func (r *partRepository) FindWithFilters(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]Part, error) {
	builder := r.psq.Select(partColumns).From("parts").Where("deleted_at IS NULL")

	for key, value := range filters {
		switch key {
		case "ids_in":
			builder = builder.Where(squirrel.Eq{"id": value})
		case "to_delete_at_is_null":
			if value.(bool) {
				builder = builder.Where("to_delete_at IS NULL")
			}
		case "quantity_gte":
			builder = builder.Where(squirrel.GtOrEq{"quantity": value})
		case "category_ilike":
			builder = builder.Where("category ILIKE ?", "%"+value.(string)+"%")
		case "brand_ilike":
			builder = builder.Where("brand ILIKE ?", "%"+value.(string)+"%")
		case "model_ilike":
			builder = builder.Where("model ILIKE ?", "%"+value.(string)+"%")
		case "location_ilike":
			builder = builder.Where("location ILIKE ?", "%"+value.(string)+"%")
		case "address_ilike":
			builder = builder.Where("address ILIKE ?", "%"+value.(string)+"%")
		case "salesman_ilike":
			builder = builder.Where("salesman ILIKE ?", "%"+value.(string)+"%")
		case "status":
			builder = builder.Where(squirrel.Eq{"status": value})
		case "has_photo":
			if value.(bool) {
				builder = builder.Where("photos IS NOT NULL AND CASE WHEN jsonb_typeof(photos) = 'array' THEN jsonb_array_length(photos) > 0 ELSE false END")
			} else {
				builder = builder.Where("(photos IS NULL OR CASE WHEN jsonb_typeof(photos) = 'array' THEN jsonb_array_length(photos) = 0 ELSE true END)")
			}
		case "number_ilike":
			builder = builder.Where("number ILIKE ?", "%"+value.(string)+"%")
		case "oem_code_ilike":
			builder = builder.Where("oem_code ILIKE ?", "%"+value.(string)+"%")
		case "vin_ilike":
			builder = builder.Where("vin ILIKE ?", "%"+value.(string)+"%")
		case "body_brand_ilike":
			builder = builder.Where("body_brand ILIKE ?", "%"+value.(string)+"%")
		case "engine_brand_ilike":
			builder = builder.Where("engine_brand ILIKE ?", "%"+value.(string)+"%")
		case "car_release_date_ilike":
			builder = builder.Where("car_release_date ILIKE ?", "%"+value.(string)+"%")
		case "car_release_period_ilike":
			builder = builder.Where("car_release_period ILIKE ?", "%"+value.(string)+"%")
		case "transmission_ilike":
			builder = builder.Where("transmission ILIKE ?", "%"+value.(string)+"%")
		case "drive_ilike":
			builder = builder.Where("drive ILIKE ?", "%"+value.(string)+"%")
		case "condition_ilike":
			builder = builder.Where("condition ILIKE ?", "%"+value.(string)+"%")
		case "manufacturer_ilike":
			builder = builder.Where("manufacturer ILIKE ?", "%"+value.(string)+"%")
		case "defect_ilike":
			builder = builder.Where("defect ILIKE ?", "%"+value.(string)+"%")
		case "color_ilike":
			builder = builder.Where("color ILIKE ?", "%"+value.(string)+"%")
		case "min_price":
			builder = builder.Where(squirrel.GtOrEq{"price": value})
		case "max_price":
			builder = builder.Where(squirrel.LtOrEq{"price": value})
		case "min_quantity":
			builder = builder.Where(squirrel.GtOrEq{"quantity": value})
		case "max_quantity":
			builder = builder.Where(squirrel.LtOrEq{"quantity": value})
		case "front_rear_ilike":
			synonyms := ExpandFrontRearSynonyms(value.(string))
			conds := make([]squirrel.Sqlizer, 0, len(synonyms))
			for _, syn := range synonyms {
				if len(syn) == 1 {
					conds = append(conds, squirrel.Eq{"front_rear": syn})
				} else {
					conds = append(conds, squirrel.Expr("front_rear ILIKE ?", "%"+syn+"%"))
				}
			}
			builder = builder.Where(squirrel.Or(conds))
		case "left_right_ilike":
			synonyms := ExpandLeftRightSynonyms(value.(string))
			conds := make([]squirrel.Sqlizer, 0, len(synonyms))
			for _, syn := range synonyms {
				if len(syn) == 1 {
					conds = append(conds, squirrel.Eq{"left_right": syn})
				} else {
					conds = append(conds, squirrel.Expr("left_right ILIKE ?", "%"+syn+"%"))
				}
			}
			builder = builder.Where(squirrel.Or(conds))
		case "top_bottom_ilike":
			synonyms := ExpandTopBottomSynonyms(value.(string))
			conds := make([]squirrel.Sqlizer, 0, len(synonyms))
			for _, syn := range synonyms {
				if len(syn) == 1 {
					conds = append(conds, squirrel.Eq{"top_bottom": syn})
				} else {
					conds = append(conds, squirrel.Expr("top_bottom ILIKE ?", "%"+syn+"%"))
				}
			}
			builder = builder.Where(squirrel.Or(conds))
		case "manufacturer_code_ilike":
			builder = builder.Where("manufacturer_code ILIKE ?", "%"+value.(string)+"%")
		case "supplier_code_ilike":
			builder = builder.Where("supplier_code ILIKE ?", "%"+value.(string)+"%")
		case "transmission_model_ilike":
			builder = builder.Where("transmission_model ILIKE ?", "%"+value.(string)+"%")
		case "wear_percentage_ilike":
			builder = builder.Where("wear_percentage ILIKE ?", "%"+value.(string)+"%")
		case "season_ilike":
			builder = builder.Where("season ILIKE ?", "%"+value.(string)+"%")
		case "diameter_ilike":
			builder = builder.Where("diameter ILIKE ?", "%"+value.(string)+"%")
		case "width_ilike":
			builder = builder.Where("width ILIKE ?", "%"+value.(string)+"%")
		case "profile_ilike":
			builder = builder.Where("profile ILIKE ?", "%"+value.(string)+"%")
		case "tire_quantity_ilike":
			builder = builder.Where("tire_quantity ILIKE ?", "%"+value.(string)+"%")
		case "drilling_ilike":
			builder = builder.Where("drilling ILIKE ?", "%"+value.(string)+"%")
		case "offset_ilike":
			builder = builder.Where("\"offset\" ILIKE ?", "%"+value.(string)+"%")
		case "center_hole_diameter_ilike":
			builder = builder.Where("center_hole_diameter ILIKE ?", "%"+value.(string)+"%")
		case "tire_model_ilike":
			builder = builder.Where("tire_model ILIKE ?", "%"+value.(string)+"%")
		case "search":
			searchStr := strings.TrimSpace(value.(string))
			if searchStr != "" {
				terms := strings.Fields(searchStr)
				for _, term := range terms {
					termPattern := "%" + term + "%"
					baseExpr := "(name ILIKE ? OR description ILIKE ? OR brand ILIKE ? OR model ILIKE ? OR number ILIKE ? OR oem_code ILIKE ? OR vin ILIKE ? OR category ILIKE ? OR car_release_date ILIKE ? OR car_release_period ILIKE ? OR body_brand ILIKE ? OR engine_brand ILIKE ? OR front_rear ILIKE ? OR left_right ILIKE ? OR top_bottom ILIKE ? OR color ILIKE ? OR condition ILIKE ? OR transmission ILIKE ? OR transmission_model ILIKE ? OR drive ILIKE ? OR defect ILIKE ? OR wear_percentage ILIKE ? OR season ILIKE ? OR diameter ILIKE ? OR width ILIKE ? OR profile ILIKE ? OR drilling ILIKE ? OR \"offset\" ILIKE ? OR center_hole_diameter ILIKE ? OR tire_model ILIKE ? OR tire_quantity ILIKE ? OR location ILIKE ? OR address ILIKE ? OR salesman ILIKE ? OR manufacturer ILIKE ? OR manufacturer_code ILIKE ? OR supplier_code ILIKE ?"
					args := make([]interface{}, 37)
					for i := range args {
						args[i] = termPattern
					}

					cleanTerm := strings.ToLower(term)
					if strings.HasPrefix(cleanTerm, "перед") || cleanTerm == "front" {
						baseExpr += " OR front_rear = 'F' OR front_rear ILIKE '%перед%'"
					} else if strings.HasPrefix(cleanTerm, "зад") || cleanTerm == "rear" {
						baseExpr += " OR front_rear = 'R' OR front_rear ILIKE '%зад%'"
					}
					if strings.HasPrefix(cleanTerm, "прав") || cleanTerm == "right" {
						baseExpr += " OR left_right = 'R' OR left_right ILIKE '%прав%'"
					} else if strings.HasPrefix(cleanTerm, "лев") || cleanTerm == "left" {
						baseExpr += " OR left_right = 'L' OR left_right ILIKE '%лев%'"
					}
					if strings.HasPrefix(cleanTerm, "верх") || cleanTerm == "top" || cleanTerm == "upper" {
						baseExpr += " OR top_bottom = 'T' OR top_bottom = 'U' OR top_bottom ILIKE '%верх%'"
					} else if strings.HasPrefix(cleanTerm, "низ") || cleanTerm == "bottom" || cleanTerm == "lower" {
						baseExpr += " OR top_bottom = 'B' OR top_bottom = 'L' OR top_bottom ILIKE '%низ%'"
					}
					baseExpr += ")"
					builder = builder.Where(squirrel.Expr(baseExpr, args...))
				}
			}
		}
	}

	builder = builder.OrderBy("id DESC")

	if limit > 0 {
		builder = builder.Limit(uint64(limit)).Offset(uint64(offset))
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanParts(rows)
}

// ─── Specific operations ────────────────────────────────────────────────────

func (r *partRepository) MarkForDeletion(ctx context.Context, id int64, deleteAt time.Time) error {
	return r.queries.MarkForDeletion(ctx, sqlc.MarkForDeletionParams{
		ID:         id,
		ToDeleteAt: pgtype.Timestamptz{Time: deleteAt, Valid: true},
	})
}

func (r *partRepository) DeleteExpiredParts(ctx context.Context, before time.Time) error {
	_, err := r.queries.DeleteExpiredParts(ctx, pgtype.Timestamptz{Time: before, Valid: true})
	return err
}

func (r *partRepository) GetStatistics(ctx context.Context) (StatisticsResponse, error) {
	var stats StatisticsResponse

	totals, err := r.queries.GetStatsTotals(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get totals")
		return stats, err
	}
	stats.TotalParts = int(totals.TotalParts)
	stats.TotalQuantity = int(totals.TotalQuantity)
	stats.TotalValue = totals.TotalValue

	categories, err := r.queries.GetStatsCategories(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get categories")
		return stats, err
	}

	catList := make([]CategoryCount, len(categories))
	for i, c := range categories {
		catList[i] = CategoryCount{
			Name:  c.Name,
			Count: int(c.Count),
		}
	}
	stats.Categories = catList

	logrus.WithFields(logrus.Fields{
		"total_parts":      stats.TotalParts,
		"total_value":      stats.TotalValue,
		"categories_count": len(stats.Categories),
	}).Info("Statistics retrieved successfully")

	return stats, nil
}

func (r *partRepository) BulkDelete(ctx context.Context, ids []int64) error {
	logrus.WithFields(logrus.Fields{
		"ids":   ids,
		"count": len(ids),
	}).Info("PartRepository.BulkDelete: Starting bulk delete")

	if err := r.queries.BulkDeleteByIDs(ctx, ids); err != nil {
		logrus.WithError(err).Error("PartRepository.BulkDelete: Failed to execute delete query")
		return err
	}

	logrus.Info("PartRepository.BulkDelete: Successfully completed bulk delete")
	return nil
}

func (r *partRepository) BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error) {
	logrus.WithField("count", len(updates)).Info("PartRepository.BulkUpdate: Starting bulk update")

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
		logrus.WithError(err).Error("PartRepository.BulkUpdate: Failed to create temp table")
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
			logrus.WithError(err).WithField("index", i).Error("PartRepository.BulkUpdate: Failed to insert into temp table")
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
		logrus.WithError(err).Error("PartRepository.BulkUpdate: Failed to execute batch update")
		return 0, err
	}

	// Количество обновленных строк
	var updatedCount int64
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM temp_parts_update").Scan(&updatedCount)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		logrus.WithError(err).Error("PartRepository.BulkUpdate: Failed to commit transaction")
		return 0, err
	}

	logrus.WithField("updated_count", updatedCount).Info("PartRepository.BulkUpdate: Successfully completed bulk update")
	return int(updatedCount), nil
}

// ─── Supplier operations ────────────────────────────────────────────────────

func (r *partRepository) DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) {
	logrus.WithField("supplier_code", supplierCode).Info("Repository: DeleteZeroQuantityPartsBySupplier")

	count, err := r.queries.CountZeroQuantityBySupplier(ctx, supplierCode)
	if err != nil {
		return 0, err
	}
	logrus.WithField("count", count).Info("Repository: Found parts to delete")

	deleted, err := r.queries.DeleteZeroQuantityBySupplier(ctx, supplierCode)
	if err != nil {
		return 0, err
	}

	logrus.WithField("deleted", deleted).Info("Repository: Successfully deleted parts")
	return deleted, nil
}

func (r *partRepository) GetSupplierCodes(ctx context.Context) ([]string, error) {
	codes, err := r.queries.GetSupplierCodes(ctx)
	if err != nil {
		return nil, err
	}
	logrus.WithField("count", len(codes)).Info("Repository: GetSupplierCodes")
	return codes, nil
}

// ─── Earnings ───────────────────────────────────────────────────────────────

func (r *partRepository) GetTotalEarnings(ctx context.Context) (float64, error) {
	earning, err := r.queries.GetEarnings(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			created, err := r.queries.CreateEarnings(ctx, 0)
			if err != nil {
				return 0, err
			}
			return created.TotalAmount, nil
		}
		return 0, err
	}
	return earning.TotalAmount, nil
}

func (r *partRepository) UpdateTotalEarnings(ctx context.Context, amount float64) error {
	earning, err := r.queries.GetEarnings(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = r.queries.CreateEarnings(ctx, amount)
			return err
		}
		return err
	}
	return r.queries.UpdateEarnings(ctx, sqlc.UpdateEarningsParams{
		TotalAmount: amount,
		ID:          earning.ID,
	})
}

// ─── Additional methods (moved from direct db access) ───────────────────────

func (r *partRepository) UpdatePartPhotos(ctx context.Context, id int64, photos StringArray) error {
	photosJSON, _ := json.Marshal(photos)
	return r.queries.UpdatePartPhotos(ctx, sqlc.UpdatePartPhotosParams{
		ID:     id,
		Photos: photosJSON,
	})
}

func (r *partRepository) GetPartsForXML(ctx context.Context) ([]Part, error) {
	parts, err := r.queries.GetPartsForXML(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]Part, len(parts))
	for i, p := range parts {
		res[i] = *sqlcPartToDomain(p)
	}
	return res, nil
}

func (r *partRepository) GetLastCreatedPart(ctx context.Context) (*Part, error) {
	p, err := r.queries.GetLastCreatedPart(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return sqlcPartToDomain(p), nil
}

// InventoryVersion возвращает отпечаток состояния склада для условных запросов.
//
// Пара «последнее изменение + число живых строк» ловит все четыре вида правок:
// вставка и обновление двигают updated_at, мягкое и жёсткое удаление меняют
// счётчик. Запрос ложится на частичный индекс idx_parts_updated_at_alive и
// стоит на порядки дешевле, чем собрать сам ответ.
func (r *partRepository) InventoryVersion(ctx context.Context) (string, error) {
	row, err := r.queries.GetInventoryVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read inventory version: %w", err)
	}

	if !row.LastChange.Valid {
		return fmt.Sprintf("empty-%d", row.Alive), nil
	}
	return fmt.Sprintf("%d-%d", row.LastChange.Time.UTC().UnixNano(), row.Alive), nil
}

// RenameSeller приводит копию имени продавца в строках запчастей в соответствие
// со справочником и возвращает обновлённые строки для переиндексации.
//
// Обновление идёт одним запросом по seller_id (индекс idx_parts_seller_id), а
// не по одной строке: у продавца могут быть тысячи позиций, и после дефектовки
// это обычное дело.
func (r *partRepository) RenameSeller(ctx context.Context, sellerID int64, name string) ([]Part, error) {
	if sellerID <= 0 || name == "" {
		return nil, nil
	}

	ids, err := r.queries.RenameSellerInParts(ctx, sqlc.RenameSellerInPartsParams{
		SellerID: sellerID,
		Salesman: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to rename seller in parts: %w", err)
	}

	if len(ids) == 0 {
		return nil, nil
	}

	updated := make([]Part, 0, len(ids))
	for _, id := range ids {
		part, err := r.FindByID(ctx, id)
		if err != nil || part == nil {
			continue
		}
		updated = append(updated, *part)
	}
	return updated, nil
}
