package service

import (
	"context"
	"fmt"

	"template-service/internal/domain"
	"template-service/internal/repository"
)

type ItemService interface {
	GetItem(ctx context.Context, id int64) (*domain.Item, error)
	ListItems(ctx context.Context, page, limit int32) ([]domain.Item, error)
	CreateItem(ctx context.Context, item *domain.Item) (*domain.Item, error)
	UpdateItem(ctx context.Context, item *domain.Item) (*domain.Item, error)
	DeleteItem(ctx context.Context, id int64) error
}

type itemService struct {
	repo repository.ItemRepository
}

func NewItemService(repo repository.ItemRepository) ItemService {
	return &itemService{
		repo: repo,
	}
}

func (s *itemService) GetItem(ctx context.Context, id int64) (*domain.Item, error) {
	if id <= 0 {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.GetByID(ctx, id)
}

func (s *itemService) ListItems(ctx context.Context, page, limit int32) ([]domain.Item, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	return s.repo.List(ctx, offset, limit)
}

func (s *itemService) CreateItem(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	if err := item.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}
	return s.repo.Create(ctx, item)
}

func (s *itemService) UpdateItem(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	if item.ID <= 0 {
		return nil, domain.ErrInvalidInput
	}
	if err := item.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}
	return s.repo.Update(ctx, item)
}

func (s *itemService) DeleteItem(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}
