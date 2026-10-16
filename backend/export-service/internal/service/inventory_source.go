package service

import (
	"context"

	"export-service/internal/domain"
)

// InventorySource определяет контракт взаимодействия со складом и продавцами
type InventorySource interface {
	InventoryVersion(ctx context.Context) (string, error)
	PartsForExport(ctx context.Context) ([]domain.Part, error)
	GetPart(ctx context.Context, partID int64) (*domain.Part, error)
	SellerINNs(ctx context.Context) (map[int64]string, error)
}
