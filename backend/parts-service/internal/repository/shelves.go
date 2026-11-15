package repository

import (
	"strings"
	"fmt"
	"context"

	"parts-service/internal/domain"
)

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
