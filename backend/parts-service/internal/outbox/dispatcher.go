package outbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/metrics"
	"avtoplaneta/pkg/outbox"
	"parts-service/internal/repository"
	"parts-service/internal/search"
)

const (
	statisticsCacheKey             = "parts:statistics"
	inventoryCacheKey              = "parts:inventory"
	inventoryCacheKeyWithoutPhotos = "parts:inventory:without_photos"
)

// Dispatcher отвечает за надежную асинхронную обработку событий transactional outbox
// в parts-service на базе универсального движка pkg/outbox.Dispatcher.
type Dispatcher struct {
	pool         *pgxpool.Pool
	repo         repository.PartRepository
	esClient     search.ElasticsearchClient
	redisClient  *redis.Client
	pollInterval time.Duration
	batchSize    int32
	inner        *outbox.Dispatcher
}

type outboxStorageAdapter struct {
	repo repository.PartRepository
}

func (s *outboxStorageAdapter) FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]outbox.Event, error) {
	events, err := s.repo.FetchPendingOutboxEvents(ctx, limit)
	if err != nil {
		return nil, err
	}
	res := make([]outbox.Event, len(events))
	for i, e := range events {
		res[i] = outbox.Event{
			ID:            e.ID,
			AggregateType: e.AggregateType,
			AggregateID:   e.AggregateID,
			EventType:     e.EventType,
			Payload:       e.Payload,
			RetryCount:    e.RetryCount,
		}
	}
	return res, nil
}

func (s *outboxStorageAdapter) MarkOutboxEventProcessed(ctx context.Context, id int64) error {
	return s.repo.MarkOutboxEventProcessed(ctx, id)
}

func (s *outboxStorageAdapter) MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error {
	return s.repo.MarkOutboxEventFailed(ctx, id, lastErr)
}

func (s *outboxStorageAdapter) CleanupProcessedOutboxEvents(ctx context.Context) error {
	return s.repo.CleanupProcessedOutboxEvents(ctx)
}

// NewDispatcher создает новый экземпляр диспетчера Outbox для parts-service
func NewDispatcher(
	pool *pgxpool.Pool,
	repo repository.PartRepository,
	esClient search.ElasticsearchClient,
	redisClient *redis.Client,
) *Dispatcher {
	return &Dispatcher{
		pool:         pool,
		repo:         repo,
		esClient:     esClient,
		redisClient:  redisClient,
		pollInterval: 10 * time.Millisecond,
		batchSize:    50,
	}
}

// Start запускает цикл диспетчера через pkg/outbox
func (d *Dispatcher) Start(ctx context.Context) error {
	storage := &outboxStorageAdapter{repo: d.repo}
	opts := outbox.Options{
		Name:         "Parts",
		Channel:      constants.OutboxChannelParts,
		BatchSize:    d.batchSize,
		PollInterval: d.pollInterval,
		Metrics:      metrics.NewOutboxMetrics("parts-service"),
	}

	d.inner = outbox.NewDispatcher(d.pool, storage, d.processEvent, opts)
	return d.inner.Start(ctx)
}

// processEvent диспетчеризирует обработку отдельного события
func (d *Dispatcher) processEvent(ctx context.Context, event outbox.Event) error {
	switch event.EventType {
	case "part_created", "part_updated", "part_photos_updated":
		return d.handlePartUpsert(ctx, event.AggregateID)
	case "part_deleted":
		return d.handlePartDelete(ctx, event.AggregateID)
	default:
		logrus.WithFields(logrus.Fields{
			"event_id":   event.ID,
			"event_type": event.EventType,
		}).Warn("Outbox dispatcher: unknown event type, skipping")
		return nil
	}
}

// handlePartUpsert синхронизирует состояние созданной или обновленной детали с ES и Redis
func (d *Dispatcher) handlePartUpsert(ctx context.Context, partID int64) error {
	part, err := d.repo.FindByID(ctx, partID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logrus.WithField("part_id", partID).Info("Part not found in DB during outbox upsert, removing from ES")
			if d.esClient != nil {
				_ = d.esClient.DeletePartFromIndex(partID)
			}
			d.invalidateCachesAndPublish(ctx, "part_delete_requested", partID)
			return nil
		}
		return fmt.Errorf("failed to fetch part %d from DB: %w", partID, err)
	}

	// 1. Синхронизируем поисковый индекс в Elasticsearch
	if d.esClient != nil {
		if constants.IsVisiblePart(int32(part.Quantity)) {
			if err := d.esClient.IndexPart(part); err != nil {
				return fmt.Errorf("failed to index part %d in Elasticsearch: %w", partID, err)
			}
		} else {
			// Запчасть списана/продана (quantity == constants.PartQuantitySold), удаляем из активного индекса
			if err := d.esClient.DeletePartFromIndex(partID); err != nil {
				return fmt.Errorf("failed to delete sold part %d from Elasticsearch: %w", partID, err)
			}
		}
	}

	// 2. Сбрасываем кэши инвентаря и статистики и отправляем уведомление в Redis Stream
	d.invalidateCachesAndPublish(ctx, "part_index_requested", partID)
	return nil
}

// handlePartDelete удаляет деталь из Elasticsearch и сбрасывает кэши
func (d *Dispatcher) handlePartDelete(ctx context.Context, partID int64) error {
	if d.esClient != nil {
		if err := d.esClient.DeletePartFromIndex(partID); err != nil {
			return fmt.Errorf("failed to delete part %d from Elasticsearch: %w", partID, err)
		}
	}

	d.invalidateCachesAndPublish(ctx, "part_delete_requested", partID)
	return nil
}

// invalidateCachesAndPublish сбрасывает кэши в Redis и публикует событие в стрим
func (d *Dispatcher) invalidateCachesAndPublish(ctx context.Context, eventType string, partID int64) {
	if d.redisClient == nil {
		return
	}

	// Инвалидация кэшей
	if err := d.redisClient.Del(ctx, inventoryCacheKey, inventoryCacheKeyWithoutPhotos, statisticsCacheKey).Err(); err != nil {
		logrus.WithError(err).Warn("Outbox dispatcher: failed to invalidate Redis caches")
	}

	// Публикация в Redis Stream events:orders для других слушателей
	err := d.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: constants.StreamOrdersEvents,
		Values: map[string]interface{}{
			"type":    eventType,
			"part_id": fmt.Sprintf("%d", partID),
		},
	}).Err()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"type":    eventType,
			"part_id": partID,
		}).Warn("Outbox dispatcher: failed to publish event to Redis stream")
	}
}
