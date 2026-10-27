package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	partsv1 "avtoplaneta/gen/parts/v1"
	"parts-service/db/sqlc"
	"parts-service/internal/catalog"
	"parts-service/internal/domain"
	"parts-service/internal/search"
)

// PartRepository определяет контракт для доступа к данным запчастей
type PartRepository interface {
	Create(ctx context.Context, part *domain.Part) error
	CreateBatch(ctx context.Context, parts []domain.Part) ([]domain.Part, error)
	FindByID(ctx context.Context, id int64) (*domain.Part, error)
	FindAll(ctx context.Context) ([]domain.Part, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	Delete(ctx context.Context, id int64) error

	DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error
	IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error

	FindWithFilters(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]domain.Part, error)
	CountWithFilters(ctx context.Context, filters map[string]interface{}) (int64, error)

	MarkForDeletion(ctx context.Context, id int64, deleteAt time.Time) error
	DeleteExpiredParts(ctx context.Context, before time.Time) error
	GetStatistics(ctx context.Context) (domain.StatisticsResponse, error)
	InventoryVersion(ctx context.Context) (string, error)
	RenameSeller(ctx context.Context, sellerID int64, name string) ([]domain.Part, error)
	BulkDelete(ctx context.Context, ids []int64) error
	BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error)
	BulkUpdateWithProto(ctx context.Context, req *partsv1.BulkUpdatePartsRequest, filterMap map[string]interface{}) ([]domain.Part, error)

	DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error)
	GetSupplierCodes(ctx context.Context) ([]string, error)
	GetNextSupplierCode(ctx context.Context) (string, error)
	PeekNextSupplierCode(ctx context.Context) (string, error)
	GetNextPartNumber(ctx context.Context) (string, error)

	GetTotalEarnings(ctx context.Context) (float64, error)
	UpdateTotalEarnings(ctx context.Context, amount float64) error

	UpdatePartPhotos(ctx context.Context, id int64, photos domain.StringArray) error
	GetPartsForXML(ctx context.Context) ([]domain.Part, error)
	GetLastCreatedPart(ctx context.Context) (*domain.Part, error)

	// Shelves
	GetShelves(ctx context.Context) ([]domain.Shelf, error)
	CreateShelf(ctx context.Context, name string) (*domain.Shelf, error)
	DeleteShelf(ctx context.Context, id int64) error

	// Outbox
	FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]sqlc.OutboxEvent, error)
	MarkOutboxEventProcessed(ctx context.Context, id int64) error
	MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error
	CleanupProcessedOutboxEvents(ctx context.Context) error
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

func scanPart(row pgx.Row) (*domain.Part, error) {
	var p domain.Part
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

func scanParts(rows pgx.Rows) ([]domain.Part, error) {
	var parts []domain.Part
	for rows.Next() {
		var p domain.Part
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
		parts = []domain.Part{}
	}
	return parts, nil
}

func sqlcPartToDomain(p sqlc.Part) *domain.Part {
	res := &domain.Part{
		PartCore: domain.PartCore{
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
		PartSpecifications: domain.PartSpecifications{
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
			Condition:         p.Condition,
			SupplierCode:      p.SupplierCode,
			Defect:            p.Defect,
			Transmission:      p.Transmission,
			TransmissionModel: p.TransmissionModel,
			Drive:             p.Drive,
			WearPercentage:    p.WearPercentage,
		},
		PartTireSpecifications: domain.PartTireSpecifications{
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

func (r *partRepository) Create(ctx context.Context, part *domain.Part) error {
	photosJSON, _ := json.Marshal(part.Photos)
	if part.Photo != "" && len(part.Photos) == 0 {
		photosJSON, _ = json.Marshal(domain.StringArray{part.Photo})
	}

	var toDeleteAt pgtype.Timestamptz
	if part.ToDeleteAt != nil {
		toDeleteAt = pgtype.Timestamptz{Time: *part.ToDeleteAt, Valid: true}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx for create: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	created, err := qtx.CreatePart(ctx, sqlc.CreatePartParams{
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

	payload, _ := json.Marshal(map[string]interface{}{"id": created.ID, "name": part.Name})
	_, err = qtx.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "part",
		AggregateID:   created.ID,
		EventType:     "part_created",
		Payload:       payload,
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *partRepository) FindByID(ctx context.Context, id int64) (*domain.Part, error) {
	p, err := r.queries.GetPartByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return sqlcPartToDomain(p), nil
}

func (r *partRepository) FindAll(ctx context.Context) ([]domain.Part, error) {
	parts, err := r.queries.GetAllParts(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Part, len(parts))
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
			if arr, ok := value.(domain.StringArray); ok {
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

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx for update: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	qtx := r.queries.WithTx(tx)
	_, err = qtx.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "part",
		AggregateID:   id,
		EventType:     "part_updated",
		Payload:       []byte("{}"),
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event for update: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *partRepository) CreateBatch(ctx context.Context, parts []domain.Part) ([]domain.Part, error) {
	if len(parts) == 0 {
		return parts, nil
	}

	batch := &pgx.Batch{}
	for i := range parts {
		part := &parts[i]
		photosJSON, _ := json.Marshal(part.Photos)
		if part.Photo != "" && len(part.Photos) == 0 {
			photosJSON, _ = json.Marshal(domain.StringArray{part.Photo})
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

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction for batch create: %w", err)
	}
	defer tx.Rollback(ctx)

	br := tx.SendBatch(ctx, batch)

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
			_ = br.Close()
			return nil, fmt.Errorf("failed to scan created part %d in batch: %w", i, err)
		}
	}

	if err := br.Close(); err != nil {
		return nil, fmt.Errorf("failed to close batch in transaction: %w", err)
	}

	qtx := r.queries.WithTx(tx)
	for i := range parts {
		payload, _ := json.Marshal(map[string]interface{}{"id": parts[i].ID, "name": parts[i].Name})
		_, err = qtx.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
			AggregateType: "part",
			AggregateID:   parts[i].ID,
			EventType:     "part_created",
			Payload:       payload,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create outbox event for batch part %d: %w", parts[i].ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit batch create transaction: %w", err)
	}

	return parts, nil
}

func (r *partRepository) DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := r.queries.WithTx(tx)
	if operationID != "" {
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
	}

	if err := q.DecreasePartQuantity(ctx, sqlc.DecreasePartQuantityParams{
		ID:     id,
		Amount: int32(amount),
	}); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]interface{}{"id": id, "amount": amount, "operation": "decrease"})
	_, err = q.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "part",
		AggregateID:   id,
		EventType:     "part_updated",
		Payload:       payload,
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *partRepository) IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := r.queries.WithTx(tx)
	if operationID != "" {
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
	}

	if err := q.IncreasePartQuantity(ctx, sqlc.IncreasePartQuantityParams{
		ID:     id,
		Amount: int32(amount),
	}); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]interface{}{"id": id, "amount": amount, "operation": "increase"})
	_, err = q.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "part",
		AggregateID:   id,
		EventType:     "part_updated",
		Payload:       payload,
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *partRepository) Delete(ctx context.Context, id int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx for delete: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	if err := qtx.DeletePart(ctx, id); err != nil {
		return err
	}

	_, err = qtx.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "part",
		AggregateID:   id,
		EventType:     "part_deleted",
		Payload:       []byte("{}"),
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event for delete: %w", err)
	}

	return tx.Commit(ctx)
}

// ─── FindWithFilters (Squirrel dynamic query) ───────────────────────────────

func (r *partRepository) buildFilterPredicates(filters map[string]interface{}) []squirrel.Sqlizer {
	var preds []squirrel.Sqlizer

	for key, value := range filters {
		switch key {
		case "ids_in":
			preds = append(preds, squirrel.Eq{"id": value})
		case "to_delete_at_is_null":
			if value.(bool) {
				preds = append(preds, squirrel.Expr("to_delete_at IS NULL"))
			}
		case "quantity_gte":
			preds = append(preds, squirrel.GtOrEq{"quantity": value})
		case "category_ilike":
			if variants, ok := value.([]string); ok {
				conds := make([]squirrel.Sqlizer, 0, len(variants))
				for _, v := range variants {
					if strings.TrimSpace(v) != "" {
						conds = append(conds, squirrel.Expr("category ILIKE ?", "%"+v+"%"))
					}
				}
				if len(conds) > 0 {
					preds = append(preds, squirrel.Or(conds))
				}
			} else if str, ok := value.(string); ok && strings.TrimSpace(str) != "" {
				preds = append(preds, squirrel.Expr("category ILIKE ?", "%"+str+"%"))
			}
		case "brand_ilike":
			preds = append(preds, squirrel.Expr("brand ILIKE ?", "%"+value.(string)+"%"))
		case "model_ilike":
			preds = append(preds, squirrel.Expr("model ILIKE ?", "%"+value.(string)+"%"))
		case "location_ilike":
			preds = append(preds, squirrel.Expr("location ILIKE ?", "%"+value.(string)+"%"))
		case "address_ilike":
			preds = append(preds, squirrel.Expr("address ILIKE ?", "%"+value.(string)+"%"))
		case "salesman_ilike":
			preds = append(preds, squirrel.Expr("salesman ILIKE ?", "%"+value.(string)+"%"))
		case "status":
			preds = append(preds, squirrel.Eq{"status": value})
		case "has_photo":
			if value.(bool) {
				preds = append(preds, squirrel.Expr("photos IS NOT NULL AND CASE WHEN jsonb_typeof(photos) = 'array' THEN jsonb_array_length(photos) > 0 ELSE false END"))
			} else {
				preds = append(preds, squirrel.Expr("(photos IS NULL OR CASE WHEN jsonb_typeof(photos) = 'array' THEN jsonb_array_length(photos) = 0 ELSE true END)"))
			}
		case "number_ilike":
			val := strings.TrimSpace(value.(string))
			cleanVal := search.CleanPartNumber(val)
			if cleanVal != "" {
				preds = append(preds, squirrel.Or{
					squirrel.Eq{"number": val},
					squirrel.Expr("REPLACE(REPLACE(LOWER(number), '-', ''), ' ', '') = ?", cleanVal),
				})
			} else {
				preds = append(preds, squirrel.Eq{"number": val})
			}
		case "oem_code_ilike":
			val := strings.TrimSpace(value.(string))
			cleanVal := search.CleanPartNumber(val)
			if cleanVal != "" {
				preds = append(preds, squirrel.Or{
					squirrel.Eq{"oem_code": val},
					squirrel.Expr("REPLACE(REPLACE(LOWER(oem_code), '-', ''), ' ', '') = ?", cleanVal),
				})
			} else {
				preds = append(preds, squirrel.Eq{"oem_code": val})
			}
		case "vin_ilike":
			preds = append(preds, squirrel.Expr("vin ILIKE ?", "%"+value.(string)+"%"))
		case "body_brand_ilike":
			preds = append(preds, squirrel.Expr("body_brand ILIKE ?", "%"+value.(string)+"%"))
		case "engine_brand_ilike":
			preds = append(preds, squirrel.Expr("engine_brand ILIKE ?", "%"+value.(string)+"%"))
		case "car_release_date_ilike":
			preds = append(preds, squirrel.Expr("car_release_date ILIKE ?", "%"+value.(string)+"%"))
		case "car_release_period_ilike":
			preds = append(preds, squirrel.Expr("car_release_period ILIKE ?", "%"+value.(string)+"%"))
		case "transmission_ilike":
			preds = append(preds, squirrel.Expr("transmission ILIKE ?", "%"+value.(string)+"%"))
		case "drive_ilike":
			preds = append(preds, squirrel.Expr("drive ILIKE ?", "%"+value.(string)+"%"))
		case "condition_ilike":
			preds = append(preds, squirrel.Expr("condition ILIKE ?", "%"+value.(string)+"%"))
		case "manufacturer_ilike":
			preds = append(preds, squirrel.Expr("manufacturer ILIKE ?", "%"+value.(string)+"%"))
		case "defect_ilike":
			preds = append(preds, squirrel.Expr("defect ILIKE ?", "%"+value.(string)+"%"))
		case "color_ilike":
			preds = append(preds, squirrel.Expr("color ILIKE ?", "%"+value.(string)+"%"))
		case "min_price":
			preds = append(preds, squirrel.GtOrEq{"price": value})
		case "max_price":
			preds = append(preds, squirrel.LtOrEq{"price": value})
		case "min_quantity":
			preds = append(preds, squirrel.GtOrEq{"quantity": value})
		case "max_quantity":
			preds = append(preds, squirrel.LtOrEq{"quantity": value})
		case "front_rear_ilike":
			synonyms := catalog.ExpandFrontRearSynonyms(value.(string))
			conds := make([]squirrel.Sqlizer, 0, len(synonyms))
			for _, syn := range synonyms {
				if len(syn) == 1 {
					conds = append(conds, squirrel.Eq{"front_rear": syn})
				} else {
					conds = append(conds, squirrel.Expr("front_rear ILIKE ?", "%"+syn+"%"))
				}
			}
			preds = append(preds, squirrel.Or(conds))
		case "left_right_ilike":
			synonyms := catalog.ExpandLeftRightSynonyms(value.(string))
			conds := make([]squirrel.Sqlizer, 0, len(synonyms))
			for _, syn := range synonyms {
				if len(syn) == 1 {
					conds = append(conds, squirrel.Eq{"left_right": syn})
				} else {
					conds = append(conds, squirrel.Expr("left_right ILIKE ?", "%"+syn+"%"))
				}
			}
			preds = append(preds, squirrel.Or(conds))
		case "top_bottom_ilike":
			synonyms := catalog.ExpandTopBottomSynonyms(value.(string))
			conds := make([]squirrel.Sqlizer, 0, len(synonyms))
			for _, syn := range synonyms {
				if len(syn) == 1 {
					conds = append(conds, squirrel.Eq{"top_bottom": syn})
				} else {
					conds = append(conds, squirrel.Expr("top_bottom ILIKE ?", "%"+syn+"%"))
				}
			}
			preds = append(preds, squirrel.Or(conds))
		case "manufacturer_code_ilike":
			val := strings.TrimSpace(value.(string))
			cleanVal := search.CleanPartNumber(val)
			if cleanVal != "" {
				preds = append(preds, squirrel.Or{
					squirrel.Eq{"manufacturer_code": val},
					squirrel.Expr("REPLACE(REPLACE(LOWER(manufacturer_code), '-', ''), ' ', '') = ?", cleanVal),
				})
			} else {
				preds = append(preds, squirrel.Eq{"manufacturer_code": val})
			}
		case "supplier_code_ilike":
			preds = append(preds, squirrel.Expr("supplier_code ILIKE ?", "%"+value.(string)+"%"))
		case "transmission_model_ilike":
			preds = append(preds, squirrel.Expr("transmission_model ILIKE ?", "%"+value.(string)+"%"))
		case "wear_percentage_ilike":
			preds = append(preds, squirrel.Expr("wear_percentage ILIKE ?", "%"+value.(string)+"%"))
		case "season_ilike":
			preds = append(preds, squirrel.Expr("season ILIKE ?", "%"+value.(string)+"%"))
		case "diameter_ilike":
			preds = append(preds, squirrel.Expr("diameter ILIKE ?", "%"+value.(string)+"%"))
		case "width_ilike":
			preds = append(preds, squirrel.Expr("width ILIKE ?", "%"+value.(string)+"%"))
		case "profile_ilike":
			preds = append(preds, squirrel.Expr("profile ILIKE ?", "%"+value.(string)+"%"))
		case "tire_quantity_ilike":
			preds = append(preds, squirrel.Expr("tire_quantity ILIKE ?", "%"+value.(string)+"%"))
		case "drilling_ilike":
			preds = append(preds, squirrel.Expr("drilling ILIKE ?", "%"+value.(string)+"%"))
		case "offset_ilike":
			preds = append(preds, squirrel.Expr("\"offset\" ILIKE ?", "%"+value.(string)+"%"))
		case "center_hole_diameter_ilike":
			preds = append(preds, squirrel.Expr("center_hole_diameter ILIKE ?", "%"+value.(string)+"%"))
		case "tire_model_ilike":
			preds = append(preds, squirrel.Expr("tire_model ILIKE ?", "%"+value.(string)+"%"))
		case "search":
			searchStr := strings.TrimSpace(value.(string))
			if searchStr != "" {
				buildTermSqlizer := func(term string) squirrel.Sqlizer {
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
					return squirrel.Expr(baseExpr, args...)
				}

				fragments := search.SplitMultiSearchQuery(searchStr)
				if len(fragments) > 1 {
					var orPreds []squirrel.Sqlizer
					for _, frag := range fragments {
						var andPreds []squirrel.Sqlizer
						for _, term := range strings.Fields(frag) {
							if strings.TrimSpace(term) != "" {
								andPreds = append(andPreds, buildTermSqlizer(term))
							}
						}
						if len(andPreds) > 0 {
							orPreds = append(orPreds, squirrel.And(andPreds))
						}
					}
					if len(orPreds) > 0 {
						preds = append(preds, squirrel.Or(orPreds))
					}
				} else if len(fragments) == 1 {
					for _, term := range strings.Fields(fragments[0]) {
						if strings.TrimSpace(term) != "" {
							preds = append(preds, buildTermSqlizer(term))
						}
					}
				}
			}
		}
	}
	return preds
}

func (r *partRepository) FindWithFilters(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]domain.Part, error) {
	builder := r.psq.Select(partColumns).From("parts").Where("deleted_at IS NULL")

	for _, pred := range r.buildFilterPredicates(filters) {
		builder = builder.Where(pred)
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

func (r *partRepository) CountWithFilters(ctx context.Context, filters map[string]interface{}) (int64, error) {
	builder := r.psq.Select("COUNT(*)").From("parts").Where("deleted_at IS NULL")

	for _, pred := range r.buildFilterPredicates(filters) {
		builder = builder.Where(pred)
	}

	sql, args, err := builder.ToSql()
	if err != nil {
		return 0, err
	}

	var count int64
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// BulkUpdateWithProto выполняет массовое обновление по спецификации Protobuf
func (r *partRepository) BulkUpdateWithProto(ctx context.Context, req *partsv1.BulkUpdatePartsRequest, filterMap map[string]interface{}) ([]domain.Part, error) {
	if req == nil {
		return nil, nil
	}

	// 1. Если указан scope по IDs (или просто непустой список IDs) - используем sqlc
	if req.Scope != nil && (req.Scope.Type == partsv1.BulkScopeType_BULK_SCOPE_TYPE_IDS || (req.Scope.Type == partsv1.BulkScopeType_BULK_SCOPE_TYPE_UNSPECIFIED && len(req.Scope.Ids) > 0)) {
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

	// 2. Если указан scope по фильтру (условиям)
	if req.Scope != nil && req.Scope.Type == partsv1.BulkScopeType_BULK_SCOPE_TYPE_FILTER {
		updateBuilder := r.psq.Update("parts").
			Where("deleted_at IS NULL").
			Where(squirrel.GtOrEq{"quantity": 0})

		if filterMap != nil {
			for _, pred := range r.buildFilterPredicates(filterMap) {
				updateBuilder = updateBuilder.Where(pred)
			}
		}

		ops := req.GetOperations()
		if ops != nil {
			if ops.Price != nil {
				switch ops.Price.Mode {
				case partsv1.PriceOperationMode_PRICE_OPERATION_MODE_PERCENT:
					updateBuilder = updateBuilder.Set("price", squirrel.Expr("ROUND((price * (1.0 + ? / 100.0))::numeric, 2)::double precision", ops.Price.Value))
				case partsv1.PriceOperationMode_PRICE_OPERATION_MODE_OFFSET:
					updateBuilder = updateBuilder.Set("price", squirrel.Expr("GREATEST(0.0, price + ?)", ops.Price.Value))
				default:
					updateBuilder = updateBuilder.Set("price", ops.Price.Value)
				}
			}
			if ops.Name != nil {
				updateBuilder = updateBuilder.Set("name", ops.GetName())
			}
			if ops.Quantity != nil {
				updateBuilder = updateBuilder.Set("quantity", ops.GetQuantity())
			}
			if ops.Description != nil {
				updateBuilder = updateBuilder.Set("description", ops.GetDescription())
			}
			if ops.Category != nil {
				updateBuilder = updateBuilder.Set("category", ops.GetCategory())
			}
			if ops.Location != nil {
				updateBuilder = updateBuilder.Set("location", ops.GetLocation())
			}
			if ops.Address != nil {
				updateBuilder = updateBuilder.Set("address", ops.GetAddress())
			}
			if ops.Salesman != nil {
				updateBuilder = updateBuilder.Set("salesman", ops.GetSalesman())
			}
			if ops.SellerId != nil {
				updateBuilder = updateBuilder.Set("seller_id", ops.GetSellerId())
			}
			if ops.Status != nil {
				updateBuilder = updateBuilder.Set("status", ops.GetStatus())
			}
			if ops.Brand != nil {
				updateBuilder = updateBuilder.Set("brand", ops.GetBrand())
			}
			if ops.Model != nil {
				updateBuilder = updateBuilder.Set("model", ops.GetModel())
			}
			if ops.Vin != nil {
				updateBuilder = updateBuilder.Set("vin", ops.GetVin())
			}
			if ops.BodyBrand != nil {
				updateBuilder = updateBuilder.Set("body_brand", ops.GetBodyBrand())
			}
			if ops.EngineBrand != nil {
				updateBuilder = updateBuilder.Set("engine_brand", ops.GetEngineBrand())
			}
			if ops.CarReleaseDate != nil {
				updateBuilder = updateBuilder.Set("car_release_date", ops.GetCarReleaseDate())
			}
			if ops.CarReleasePeriod != nil {
				updateBuilder = updateBuilder.Set("car_release_period", ops.GetCarReleasePeriod())
			}
			if ops.FrontRear != nil {
				updateBuilder = updateBuilder.Set("front_rear", ops.GetFrontRear())
			}
			if ops.LeftRight != nil {
				updateBuilder = updateBuilder.Set("left_right", ops.GetLeftRight())
			}
			if ops.TopBottom != nil {
				updateBuilder = updateBuilder.Set("top_bottom", ops.GetTopBottom())
			}
			if ops.Number != nil {
				updateBuilder = updateBuilder.Set("number", ops.GetNumber())
			}
			if ops.Manufacturer != nil {
				updateBuilder = updateBuilder.Set("manufacturer", ops.GetManufacturer())
			}
			if ops.ManufacturerCode != nil {
				updateBuilder = updateBuilder.Set("manufacturer_code", ops.GetManufacturerCode())
			}
			if ops.OemCode != nil {
				updateBuilder = updateBuilder.Set("oem_code", ops.GetOemCode())
			}
			if ops.Color != nil {
				updateBuilder = updateBuilder.Set("color", ops.GetColor())
			}
			if ops.Condition != nil {
				updateBuilder = updateBuilder.Set("condition", ops.GetCondition())
			}
			if ops.SupplierCode != nil {
				updateBuilder = updateBuilder.Set("supplier_code", ops.GetSupplierCode())
			}
			if ops.Defect != nil {
				updateBuilder = updateBuilder.Set("defect", ops.GetDefect())
			}
			if ops.Transmission != nil {
				updateBuilder = updateBuilder.Set("transmission", ops.GetTransmission())
			}
			if ops.TransmissionModel != nil {
				updateBuilder = updateBuilder.Set("transmission_model", ops.GetTransmissionModel())
			}
			if ops.Drive != nil {
				updateBuilder = updateBuilder.Set("drive", ops.GetDrive())
			}
			if ops.WearPercentage != nil {
				updateBuilder = updateBuilder.Set("wear_percentage", ops.GetWearPercentage())
			}
			if ops.Season != nil {
				updateBuilder = updateBuilder.Set("season", ops.GetSeason())
			}
			if ops.Diameter != nil {
				updateBuilder = updateBuilder.Set("diameter", ops.GetDiameter())
			}
			if ops.Width != nil {
				updateBuilder = updateBuilder.Set("width", ops.GetWidth())
			}
			if ops.Profile != nil {
				updateBuilder = updateBuilder.Set("profile", ops.GetProfile())
			}
			if ops.TireQuantity != nil {
				updateBuilder = updateBuilder.Set("tire_quantity", ops.GetTireQuantity())
			}
			if ops.Drilling != nil {
				updateBuilder = updateBuilder.Set("drilling", ops.GetDrilling())
			}
			if ops.Offset != nil {
				updateBuilder = updateBuilder.Set(`"offset"`, ops.GetOffset())
			}
			if ops.CenterHoleDiameter != nil {
				updateBuilder = updateBuilder.Set("center_hole_diameter", ops.GetCenterHoleDiameter())
			}
			if ops.TireModel != nil {
				updateBuilder = updateBuilder.Set("tire_model", ops.GetTireModel())
			}
		}

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

	return []domain.Part{}, nil
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

func calculateGrowth(current, previous float64) float64 {
	if previous <= 0 {
		return 0
	}
	growth := ((current - previous) / previous) * 100.0
	return math.Round(growth*10) / 10
}

func (r *partRepository) GetStatistics(ctx context.Context) (domain.StatisticsResponse, error) {
	var stats domain.StatisticsResponse

	totals, err := r.queries.GetStatsTotals(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get totals")
		return stats, err
	}
	stats.TotalParts = int(totals.TotalParts)
	stats.TotalQuantity = int(totals.TotalQuantity)
	stats.TotalValue = totals.TotalValue

	now := time.Now()
	// Срез на начало текущего месяца
	cutoff := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	prev, err := r.queries.GetStatsTotalsBefore(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
	if err == nil && prev.PrevParts == 0 {
		// Fallback: 30 дней назад, если на начало месяца данных не было
		cutoff = now.AddDate(0, 0, -30)
		prev, err = r.queries.GetStatsTotalsBefore(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})
	}

	if err == nil && prev.PrevParts > 0 {
		stats.PartsGrowth = calculateGrowth(float64(totals.TotalParts), float64(prev.PrevParts))
		stats.QuantityGrowth = calculateGrowth(float64(totals.TotalQuantity), float64(prev.PrevQuantity))
		stats.ValueGrowth = calculateGrowth(totals.TotalValue, prev.PrevValue)
	}

	categories, err := r.queries.GetStatsCategories(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get categories")
		return stats, err
	}

	catList := make([]domain.CategoryCount, len(categories))
	for i, c := range categories {
		catList[i] = domain.CategoryCount{
			Name:  c.Name,
			Count: int(c.Count),
		}
	}
	stats.Categories = catList

	logrus.WithFields(logrus.Fields{
		"total_parts":      stats.TotalParts,
		"total_value":      stats.TotalValue,
		"parts_growth":     stats.PartsGrowth,
		"quantity_growth":  stats.QuantityGrowth,
		"value_growth":     stats.ValueGrowth,
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

func (r *partRepository) DeleteZeroQuantityPartsWithIDs(ctx context.Context, supplierCode string) ([]int64, error) {
	supplierCode = strings.TrimSpace(supplierCode)
	logrus.WithField("supplier_code", supplierCode).Info("Repository: DeleteZeroQuantityPartsWithIDs")

	var (
		deletedIDs []int64
		err        error
	)

	if strings.HasPrefix(supplierCode, "car:") {
		raw := strings.TrimPrefix(supplierCode, "car:")
		parts := strings.SplitN(raw, "|", 4)
		for len(parts) < 4 {
			parts = append(parts, "")
		}
		deletedIDs, err = r.queries.DeleteZeroQuantityPartsByCar(ctx, sqlc.DeleteZeroQuantityPartsByCarParams{
			Brand:          parts[0],
			Model:          parts[1],
			CarReleaseDate: parts[2],
			Vin:            parts[3],
		})
	} else {
		deletedIDs, err = r.queries.DeleteZeroQuantityPartsBySupplierCode(ctx, supplierCode)
	}
	if err != nil {
		return nil, err
	}

	logrus.WithField("deleted", len(deletedIDs)).Info("Repository: Successfully deleted zero-quantity parts")
	return deletedIDs, nil
}

func (r *partRepository) DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) {
	ids, err := r.DeleteZeroQuantityPartsWithIDs(ctx, supplierCode)
	if err != nil {
		return 0, err
	}
	return int64(len(ids)), nil
}

func (r *partRepository) GetSupplierBatches(ctx context.Context) ([]domain.SupplierBatchInfo, error) {
	rows, err := r.queries.GetSupplierBatches(ctx)
	if err != nil {
		logrus.WithError(err).Error("Repository: GetSupplierBatches query failed")
		return nil, err
	}

	var batches []domain.SupplierBatchInfo
	for _, row := range rows {
		var code string
		switch v := row.BatchCode.(type) {
		case string:
			code = v
		case []byte:
			code = string(v)
		default:
			code = fmt.Sprintf("%v", v)
		}

		b := domain.SupplierBatchInfo{
			Code:      code,
			Brand:     row.BrandName,
			Model:     row.ModelName,
			Year:      row.ReleaseDate,
			VIN:       row.CarVin,
			ZeroCount: row.ZeroCount,
		}

		carTitle := strings.TrimSpace(strings.Join([]string{b.Brand, b.Model, b.Year}, " "))
		if carTitle == "" {
			carTitle = "Без марки/модели"
		}
		if b.VIN != "" {
			carTitle += fmt.Sprintf(" (VIN: %s)", b.VIN)
		}
		if strings.HasPrefix(b.Code, "car:") {
			b.Label = fmt.Sprintf("%s — %d шт. с 0 [без кода поставки]", carTitle, b.ZeroCount)
		} else {
			b.Label = fmt.Sprintf("%s — %d шт. с 0 [код: %s]", carTitle, b.ZeroCount, b.Code)
		}
		batches = append(batches, b)
	}

	logrus.WithField("count", len(batches)).Info("Repository: GetSupplierBatches loaded batches")
	return batches, nil
}

func (r *partRepository) GetSupplierCodes(ctx context.Context) ([]string, error) {
	batches, err := r.GetSupplierBatches(ctx)
	if err == nil {
		codes := make([]string, 0, len(batches))
		for _, b := range batches {
			codes = append(codes, b.Code)
		}
		logrus.WithField("count", len(codes)).Info("Repository: GetSupplierCodes from batches")
		return codes, nil
	}

	logrus.WithError(err).Warn("Repository: GetSupplierBatches failed, falling back to GetSupplierCodes query")
	codes, err := r.queries.GetSupplierCodes(ctx)
	if err != nil {
		logrus.WithError(err).Error("Repository: GetSupplierCodes query failed")
		return nil, err
	}
	logrus.WithField("count", len(codes)).Info("Repository: GetSupplierCodes")
	return codes, nil
}

func (r *partRepository) GetNextSupplierCode(ctx context.Context) (string, error) {
	seqVal, err := r.queries.GetNextDefectBatchSequence(ctx)
	if err != nil {
		logrus.WithError(err).Warn("Repository: nextval(defect_batch_seq) failed, falling back to max supplier_code query")
		maxVal, fallbackErr := r.queries.GetMaxSupplierCodeFallback(ctx)
		if fallbackErr == nil && maxVal > 0 {
			return strconv.FormatInt(int64(maxVal), 10), nil
		}
		return "1", nil
	}
	return strconv.FormatInt(seqVal, 10), nil
}

func (r *partRepository) PeekNextSupplierCode(ctx context.Context) (string, error) {
	var lastVal int64
	var isCalled bool
	err := r.pool.QueryRow(ctx, "SELECT last_value, is_called FROM defect_batch_seq").Scan(&lastVal, &isCalled)
	if err != nil {
		maxVal, fallbackErr := r.queries.GetMaxSupplierCodeFallback(ctx)
		if fallbackErr == nil && maxVal > 0 {
			return strconv.FormatInt(int64(maxVal), 10), nil
		}
		return "1", nil
	}
	if !isCalled {
		return strconv.FormatInt(lastVal, 10), nil
	}
	return strconv.FormatInt(lastVal+1, 10), nil
}

func (r *partRepository) GetNextPartNumber(ctx context.Context) (string, error) {
	seqVal, err := r.queries.GetNextPartNumberSequence(ctx)
	if err != nil {
		logrus.WithError(err).Warn("Repository: nextval(part_number_seq) failed, falling back to max number query")
		maxVal, fallbackErr := r.queries.GetMaxPartNumberFallback(ctx)
		if fallbackErr == nil && maxVal > 0 {
			return strconv.FormatInt(int64(maxVal), 10), nil
		}
		return "1", nil
	}
	return strconv.FormatInt(seqVal, 10), nil
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

func (r *partRepository) UpdatePartPhotos(ctx context.Context, id int64, photos domain.StringArray) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx for update photos: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	photosJSON, _ := json.Marshal(photos)
	err = qtx.UpdatePartPhotos(ctx, sqlc.UpdatePartPhotosParams{
		ID:     id,
		Photos: photosJSON,
	})
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]interface{}{"id": id, "photos": photos})
	_, err = qtx.CreateOutboxEvent(ctx, sqlc.CreateOutboxEventParams{
		AggregateType: "part",
		AggregateID:   id,
		EventType:     "part_photos_updated",
		Payload:       payload,
	})
	if err != nil {
		return fmt.Errorf("failed to create outbox event for photos: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *partRepository) GetPartsForXML(ctx context.Context) ([]domain.Part, error) {
	parts, err := r.queries.GetPartsForXML(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Part, len(parts))
	for i, p := range parts {
		res[i] = *sqlcPartToDomain(p)
	}
	return res, nil
}

func (r *partRepository) GetLastCreatedPart(ctx context.Context) (*domain.Part, error) {
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
func (r *partRepository) RenameSeller(ctx context.Context, sellerID int64, name string) ([]domain.Part, error) {
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

	updated := make([]domain.Part, 0, len(ids))
	for _, id := range ids {
		part, err := r.FindByID(ctx, id)
		if err != nil || part == nil {
			continue
		}
		updated = append(updated, *part)
	}
	return updated, nil
}

// ─── Outbox ─────────────────────────────────────────────────────────────────

func (r *partRepository) FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]sqlc.OutboxEvent, error) {
	return r.queries.FetchPendingOutboxEvents(ctx, limit)
}

func (r *partRepository) MarkOutboxEventProcessed(ctx context.Context, id int64) error {
	return r.queries.MarkOutboxEventProcessed(ctx, id)
}

func (r *partRepository) MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error {
	return r.queries.MarkOutboxEventFailed(ctx, sqlc.MarkOutboxEventFailedParams{
		ID:        id,
		LastError: pgtype.Text{String: lastErr, Valid: lastErr != ""},
	})
}

func (r *partRepository) CleanupProcessedOutboxEvents(ctx context.Context) error {
	return r.queries.CleanupProcessedOutboxEvents(ctx)
}

// ─── Shelves ────────────────────────────────────────────────────────────────

func (r *partRepository) GetShelves(ctx context.Context) ([]domain.Shelf, error) {
	shelves, err := r.queries.ListShelves(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list shelves: %w", err)
	}
	result := make([]domain.Shelf, len(shelves))
	for i, s := range shelves {
		result[i] = domain.Shelf{
			ID:        s.ID,
			Name:      s.Name,
			CreatedAt: s.CreatedAt.Time,
		}
	}
	return result, nil
}

func (r *partRepository) CreateShelf(ctx context.Context, name string) (*domain.Shelf, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, fmt.Errorf("shelf name cannot be empty")
	}
	shelf, err := r.queries.CreateShelf(ctx, trimmed)
	if err != nil {
		return nil, fmt.Errorf("failed to create shelf: %w", err)
	}
	return &domain.Shelf{
		ID:        shelf.ID,
		Name:      shelf.Name,
		CreatedAt: shelf.CreatedAt.Time,
	}, nil
}

func (r *partRepository) DeleteShelf(ctx context.Context, id int64) error {
	rows, err := r.queries.DeleteShelf(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete shelf: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("shelf not found")
	}
	return nil
}

