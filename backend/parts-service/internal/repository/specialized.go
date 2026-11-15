package repository

import (
	"github.com/jackc/pgx/v5"
	"errors"
	"context"
	"encoding/json"
	"fmt"

	"parts-service/db/sqlc"
	"parts-service/internal/domain"
)

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
