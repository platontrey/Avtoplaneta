package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/logger"
	"parts-service/internal/domain"
	"parts-service/internal/search"
)

// AddPart добавляет новую запчасть
func (s *inventoryService) AddPart(ctx context.Context, part *domain.Part) (*domain.Part, error) {
	if err := domain.ValidatePart(part); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, part); err != nil {
		return nil, err
	}

	// Получаем созданную запчасть
	createdPart, err := s.repo.FindByID(ctx, part.ID)
	if err != nil {
		return nil, err
	}

	// Отправляем событие индексации в Redis Stream
	err = s.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: constants.StreamOrdersEvents,
		Values: map[string]interface{}{
			"type":    "part_index_requested",
			"part_id": strconv.FormatInt(createdPart.ID, 10),
		},
	}).Err()
	if err != nil {
		slog.WarnContext(ctx, "Failed to publish part_index_requested to Redis", logger.Err(err))
	}

	// Инвалидируем кэш статистики
	s.invalidateStatisticsCache(ctx)

	return createdPart, nil
}

// DuplicatePart создает дубликат существующей запчасти без фото
func (s *inventoryService) DuplicatePart(ctx context.Context, id int64) (*domain.Part, error) {
	source, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("запчасть не найдена: %w", err)
	}

	clone := *source
	clone.ID = 0
	clone.Number = ""
	clone.Photos = domain.StringArray{}
	clone.Photo = ""
	clone.ToDeleteAt = nil
	if clone.Quantity < 0 {
		clone.Quantity = 1
	}

	if err := s.repo.Create(ctx, &clone); err != nil {
		return nil, fmt.Errorf("не удалось создать дубликат запчасти: %w", err)
	}

	created, err := s.repo.FindByID(ctx, clone.ID)
	if err != nil {
		return &clone, nil
	}

	// Индексируем в ES
	if err := search.IndexPart(created); err != nil {
		slog.WarnContext(ctx, "DuplicatePart: failed to index duplicated part in Elasticsearch", logger.Err(err))
	}

	// Сброс кэша
	if s.redis != nil {
		s.redis.Del(ctx, inventoryCacheKey)
		s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
		s.invalidateStatisticsCache(ctx)
	}

	return created, nil
}

// AddPartsBatch пакетно валидирует, создает и индексирует запчасти (для дефектных ведомостей)
func (s *inventoryService) AddPartsBatch(ctx context.Context, parts []domain.Part) ([]domain.Part, error) {
	if len(parts) == 0 {
		return nil, nil
	}

	for i := range parts {
		if err := domain.ValidatePart(&parts[i]); err != nil {
			return nil, fmt.Errorf("ошибка валидации детали #%d (%s): %w", i, parts[i].Name, err)
		}
	}

	createdParts, err := s.repo.CreateBatch(ctx, parts)
	if err != nil {
		return nil, fmt.Errorf("ошибка пакетной вставки запчастей: %w", err)
	}

	// Пакетная индексация всех созданных деталей в Elasticsearch
	if err := search.BulkIndexParts(ctx, createdParts); err != nil {
		slog.WarnContext(ctx, "Не удалось пакетно проиндексировать созданные запчасти в Elasticsearch", logger.Err(err))
	}

	// Сброс кэша один раз на весь батч
	s.redis.Del(ctx, inventoryCacheKey)
	s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
	s.invalidateStatisticsCache(ctx)

	return createdParts, nil
}

// UpdatePart обновляет запчасть
func (s *inventoryService) UpdatePart(ctx context.Context, id int64, updates map[string]interface{}) error {
	// Получаем существующую часть для обработки обновлений
	existingPart, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	processedUpdates, err := domain.ProcessPartUpdates(updates, existingPart)
	if err != nil {
		return err
	}

	if err := s.repo.Update(ctx, id, processedUpdates); err != nil {
		return err
	}

	// Отправляем событие индексации в Redis Stream
	err = s.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: constants.StreamOrdersEvents,
		Values: map[string]interface{}{
			"type":    "part_index_requested",
			"part_id": strconv.FormatInt(id, 10),
		},
	}).Err()
	if err != nil {
		slog.WarnContext(ctx, "Failed to publish part_index_requested (update) to Redis", logger.Err(err))
	}

	// Инвалидируем кэш статистики
	s.invalidateStatisticsCache(ctx)

	return nil
}

// DeletePart удаляет запчасть
func (s *inventoryService) DeletePart(ctx context.Context, id int64) error {
	part, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Удаляем все фото если есть
	for _, photoPath := range part.Photos {
		if photoPath != "" {
			if err := DeletePhotoFile(photoPath); err != nil {
				slog.WarnContext(ctx, "Failed to delete photo file", slog.String("path", photoPath), logger.Err(err))
			}
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Отправляем событие удаления из индекса в Redis Stream
	err = s.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: constants.StreamOrdersEvents,
		Values: map[string]interface{}{
			"type":    "part_delete_requested",
			"part_id": strconv.FormatInt(id, 10),
		},
	}).Err()
	if err != nil {
		slog.WarnContext(ctx, "Failed to publish part_delete_requested to Redis", logger.Err(err))
	}

	// Инвалидируем кэш статистики
	s.invalidateStatisticsCache(ctx)

	return nil
}

// MarkPartForDeletion отмечает запчасть для удаления
func (s *inventoryService) MarkPartForDeletion(ctx context.Context, id int64) error {
	fourteenDaysFromNow := time.Now().AddDate(0, 0, 14)
	return s.repo.MarkForDeletion(ctx, id, fourteenDaysFromNow)
}

// GetPartByID получает запчасть по ID
func (s *inventoryService) GetPartByID(ctx context.Context, id int64) (*domain.Part, error) {
	part, err := s.repo.FindByID(ctx, id)
	if err != nil || part == nil {
		return part, err
	}

	return part, nil
}

func (s *inventoryService) DecreasePartQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	// Сброс кэша
	s.redis.Del(ctx, inventoryCacheKey)
	s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
	s.invalidateStatisticsCache(ctx)

	err := s.repo.DecreaseQuantity(ctx, id, amount, operationID)
	if err != nil {
		return err
	}

	part, err := s.GetPartByID(ctx, id)
	if err == nil && s.es != nil {
		s.es.IndexPart(part)
	}

	return nil
}

func (s *inventoryService) IncreasePartQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	// Сброс кэша
	s.redis.Del(ctx, inventoryCacheKey)
	s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
	s.invalidateStatisticsCache(ctx)

	err := s.repo.IncreaseQuantity(ctx, id, amount, operationID)
	if err != nil {
		return err
	}

	part, err := s.GetPartByID(ctx, id)
	if err == nil && s.es != nil {
		s.es.IndexPart(part)
	}

	return nil
}
