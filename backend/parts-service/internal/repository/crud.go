package repository

import (
	"errors"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"parts-service/db/sqlc"
	"parts-service/internal/domain"
)

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
			slog.InfoContext(ctx, "Stock operation already recorded, skipping decrease (idempotent)",
				slog.String("operation_id", operationID),
				slog.Int64("part_id", id))
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
			slog.InfoContext(ctx, "Stock operation already recorded, skipping increase (idempotent)",
				slog.String("operation_id", operationID),
				slog.Int64("part_id", id))
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
