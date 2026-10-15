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

	"parts-service/db/sqlc"
	"parts-service/internal/metrics"
	"parts-service/internal/repository"
	"parts-service/internal/search"
)

const (
	defaultPollInterval = 100 * time.Millisecond
	fallbackPollInterval = 3 * time.Second
	defaultBatchSize    = 50
	cleanupInterval     = 1 * time.Hour
	outboxChannel       = "outbox_events_channel"

	statisticsCacheKey             = "parts:statistics"
	inventoryCacheKey              = "parts:inventory"
	inventoryCacheKeyWithoutPhotos = "parts:inventory:without_photos"
	eventsOrdersStream             = "events:orders"
)

// Dispatcher отвечает за надежную асинхронную обработку событий из transactional outbox таблицы
// и синхронизацию состояния БД с Elasticsearch и Redis.
// Поддерживает sub-millisecond пробуждение через PostgreSQL LISTEN / NOTIFY.
type Dispatcher struct {
	pool         *pgxpool.Pool
	repo         repository.PartRepository
	esClient     search.ElasticsearchClient
	redisClient  *redis.Client
	pollInterval time.Duration
	batchSize    int32
}

// NewDispatcher создает новый экземпляр диспетчера Outbox
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
		pollInterval: defaultPollInterval,
		batchSize:    defaultBatchSize,
	}
}

// Start запускает цикл диспетчера и фоновую очистку обработанных событий.
// Если доступен pool подключений БД, подписывается на PostgreSQL LISTEN outbox_events_channel
// для мгновенной реакции на коммит транзакции.
func (d *Dispatcher) Start(ctx context.Context) error {
	logrus.Info("Starting Transactional Outbox dispatcher")

	// Фоновый цикл очистки устаревших обработанных событий
	go d.startCleanupLoop(ctx)

	notifyCh := make(chan struct{}, 1)

	// Запуск LISTEN слушателя только если передан пул БД
	if d.pool != nil {
		go d.startNotifyListener(ctx, notifyCh)
	}

	// Если работает LISTEN / NOTIFY, опрос базы служит лишь страховкой (раз в 3 сек)
	interval := d.pollInterval
	if d.pool != nil {
		interval = fallbackPollInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Первоначальная проверка при старте
	d.processPendingEvents(ctx)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping Transactional Outbox dispatcher")
			return nil

		case <-notifyCh:
			// Мгновенное пробуждение от коммита транзакции в PostgreSQL
			d.processPendingEvents(ctx)

		case <-ticker.C:
			// Периодическая проверка на случай пропущенных уведомлений
			d.processPendingEvents(ctx)
		}
	}
}

// startNotifyListener подключается к PostgreSQL и слушает канал outbox_events_channel
func (d *Dispatcher) startNotifyListener(ctx context.Context, notifyCh chan<- struct{}) {
	for {
		if ctx.Err() != nil {
			return
		}

		conn, err := d.pool.Acquire(ctx)
		if err != nil {
			logrus.WithError(err).Warn("Outbox dispatcher: failed to acquire connection for LISTEN, retrying in 2s")
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		_, err = conn.Exec(ctx, "LISTEN "+outboxChannel)
		if err != nil {
			conn.Release()
			logrus.WithError(err).Warn("Outbox dispatcher: failed to execute LISTEN, retrying in 2s")
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		logrus.Infof("Outbox dispatcher: successfully subscribed to %s via PostgreSQL LISTEN", outboxChannel)

		// Цикл ожидания уведомлений от PostgreSQL
		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				conn.Release()
				if ctx.Err() != nil {
					return
				}
				logrus.WithError(err).Warn("Outbox dispatcher: connection lost while waiting for notification, reconnecting...")
				break
			}

			if notification != nil {
				select {
				case notifyCh <- struct{}{}:
				default:
					// В канале уже есть сигнал пробуждения
				}
			}
		}
	}
}

// processPendingEvents выбирает пачку событий и обрабатывает их
func (d *Dispatcher) processPendingEvents(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		events, err := d.repo.FetchPendingOutboxEvents(ctx, d.batchSize)
		if err != nil {
			logrus.WithError(err).Error("Outbox dispatcher: failed to fetch pending events")
			return
		}

		if len(events) == 0 {
			// Очередь пуста
			return
		}

		for _, event := range events {
			if ctx.Err() != nil {
				return
			}

			start := time.Now()
			err := d.processEvent(ctx, event)
			duration := time.Since(start).Seconds()

			if err != nil {
				metrics.OutboxEventsFailedTotal.WithLabelValues(event.EventType).Inc()
				metrics.OutboxProcessingDuration.WithLabelValues(event.EventType).Observe(duration)

				logrus.WithError(err).WithFields(logrus.Fields{
					"event_id":       event.ID,
					"event_type":     event.EventType,
					"aggregate_id":   event.AggregateID,
					"aggregate_type": event.AggregateType,
				}).Error("Outbox dispatcher: failed to process event")

				if markErr := d.repo.MarkOutboxEventFailed(ctx, event.ID, err.Error()); markErr != nil {
					logrus.WithError(markErr).WithField("event_id", event.ID).Error("Outbox dispatcher: failed to mark event as failed")
				}
			} else {
				metrics.OutboxEventsProcessedTotal.WithLabelValues(event.EventType).Inc()
				metrics.OutboxProcessingDuration.WithLabelValues(event.EventType).Observe(duration)

				if markErr := d.repo.MarkOutboxEventProcessed(ctx, event.ID); markErr != nil {
					logrus.WithError(markErr).WithField("event_id", event.ID).Error("Outbox dispatcher: failed to mark event as processed")
				}
			}
		}

		// Если обработали полную пачку, сразу проверяем следующую, не дожидаясь тика
		if len(events) < int(d.batchSize) {
			return
		}
	}
}

// processEvent обрабатывает конкретное событие outbox
func (d *Dispatcher) processEvent(ctx context.Context, event sqlc.OutboxEvent) error {
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
			// Запчасть была удалена из БД до завершения обработки события создания/обновления
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
		if part.Quantity >= 0 {
			if err := d.esClient.IndexPart(part); err != nil {
				return fmt.Errorf("failed to index part %d in Elasticsearch: %w", partID, err)
			}
		} else {
			// Запчасть списана/продана (quantity == -1), удаляем из активного индекса
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
		Stream: eventsOrdersStream,
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

// startCleanupLoop периодически удаляет старые обработанные события outbox
func (d *Dispatcher) startCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.repo.CleanupProcessedOutboxEvents(ctx); err != nil {
				logrus.WithError(err).Error("Outbox dispatcher: failed to cleanup processed outbox events")
			} else {
				logrus.Debug("Outbox dispatcher: cleaned up old processed outbox events")
			}
		}
	}
}
