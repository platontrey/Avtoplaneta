package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"template-service/db/sqlc"
	"template-service/internal/domain"
)

type ItemRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Item, error)
	List(ctx context.Context, offset, limit int32) ([]domain.Item, error)
	Create(ctx context.Context, item *domain.Item) (*domain.Item, error)
	Update(ctx context.Context, item *domain.Item) (*domain.Item, error)
	Delete(ctx context.Context, id int64) error
}

type itemRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewItemRepository(pool *pgxpool.Pool) ItemRepository {
	return &itemRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *itemRepository) GetByID(ctx context.Context, id int64) (*domain.Item, error) {
	row, err := r.queries.GetItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrItemNotFound
		}
		return nil, fmt.Errorf("ошибка получения элемента по ID: %w", err)
	}
	return mapSQLCToDomain(row), nil
}

func (r *itemRepository) List(ctx context.Context, offset, limit int32) ([]domain.Item, error) {
	rows, err := r.queries.ListItems(ctx, sqlc.ListItemsParams{
		OffsetCount: offset,
		LimitCount:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка выборки списка элементов: %w", err)
	}

	items := make([]domain.Item, len(rows))
	for i, row := range rows {
		items[i] = *mapSQLCToDomain(row)
	}
	return items, nil
}

func (r *itemRepository) Create(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	metaBytes, err := json.Marshal(item.Metadata)
	if err != nil {
		metaBytes = []byte("{}")
	}

	created, err := r.queries.CreateItem(ctx, sqlc.CreateItemParams{
		Name:        item.Name,
		Description: item.Description,
		Status:      item.Status,
		Metadata:    metaBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания элемента: %w", err)
	}

	return mapSQLCToDomain(created), nil
}

func (r *itemRepository) Update(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	metaBytes, err := json.Marshal(item.Metadata)
	if err != nil {
		metaBytes = []byte("{}")
	}

	updated, err := r.queries.UpdateItem(ctx, sqlc.UpdateItemParams{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Status:      item.Status,
		Metadata:    metaBytes,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrItemNotFound
		}
		return nil, fmt.Errorf("ошибка обновления элемента: %w", err)
	}

	return mapSQLCToDomain(updated), nil
}

func (r *itemRepository) Delete(ctx context.Context, id int64) error {
	affected, err := r.queries.SoftDeleteItem(ctx, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления элемента: %w", err)
	}
	if affected == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

func mapSQLCToDomain(row sqlc.Item) *domain.Item {
	var meta map[string]any
	if len(row.Metadata) > 0 {
		_ = json.Unmarshal(row.Metadata, &meta)
	}
	if meta == nil {
		meta = make(map[string]any)
	}

	item := &domain.Item{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Status:      row.Status,
		Metadata:    meta,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}

	if row.DeletedAt.Valid {
		item.DeletedAt = &row.DeletedAt.Time
	}

	return item
}
