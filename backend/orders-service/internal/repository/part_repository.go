package repository

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"orders-service/internal/domain"
)

// PartsGRPCClient интерфейс для работы с запчастями по gRPC
type PartsGRPCClient interface {
	GetPartByID(ctx context.Context, id int64) (*domain.Part, error)
	DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error
	IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error
	DeletePart(ctx context.Context, id int64) error
	Close() error
}

// PartRepositoryForOrders определяет интерфейс для работы с запчастями (для orders-service)
type PartRepositoryForOrders interface {
	FindByID(ctx context.Context, id int64) (*domain.Part, error)
	UpdateQuantity(ctx context.Context, id int64, newQuantity int) error
	DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error
	IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error
	DeletePart(ctx context.Context, id int64) error
}

// partRepositoryForOrders реализует PartRepositoryForOrders через gRPC
type partRepositoryForOrders struct {
	client PartsGRPCClient
}

// NewPartRepositoryForOrders создает новый репозиторий деталей через gRPC
func NewPartRepositoryForOrders(client PartsGRPCClient) PartRepositoryForOrders {
	return &partRepositoryForOrders{
		client: client,
	}
}

func (r *partRepositoryForOrders) FindByID(ctx context.Context, id int64) (*domain.Part, error) {
	p, err := r.client.GetPartByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find part with ID %d via gRPC: %w", id, err)
	}
	return p, nil
}

func (r *partRepositoryForOrders) UpdateQuantity(ctx context.Context, id int64, newQuantity int) error {
	return fmt.Errorf("UpdateQuantity is not supported via gRPC yet")
}

func (r *partRepositoryForOrders) DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	err := r.client.DecreaseQuantity(ctx, id, amount, operationID)
	if err != nil {
		return fmt.Errorf("failed to decrease quantity for part %d by %d: %w", id, amount, err)
	}
	return nil
}

func (r *partRepositoryForOrders) IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	err := r.client.IncreaseQuantity(ctx, id, amount, operationID)
	if err != nil {
		return fmt.Errorf("failed to increase quantity for part %d by %d: %w", id, amount, err)
	}
	return nil
}

func (r *partRepositoryForOrders) DeletePart(ctx context.Context, id int64) error {
	logrus.WithField("part_id", id).Info("Starting part deletion via gRPC")
	err := r.client.DeletePart(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete part with ID %d via gRPC: %w", id, err)
	}
	return nil
}
