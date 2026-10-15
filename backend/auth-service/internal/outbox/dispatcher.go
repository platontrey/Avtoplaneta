package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"auth-service/db/sqlc"
	"auth-service/internal/metrics"
	"auth-service/internal/repository"
)

const (
	defaultPollInterval  = 100 * time.Millisecond
	fallbackPollInterval = 3 * time.Second
	defaultBatchSize     = 50
	cleanupInterval      = 1 * time.Hour
	outboxChannel        = "auth_outbox_channel"
	eventsOrdersStream   = "events:orders"
)

// Dispatcher отвечает за надежную асинхронную обработку событий из transactional outbox таблицы
// в auth-service и синхронизацию состояния (seller_renamed) с Redis Streams.
type Dispatcher struct {
	pool         *pgxpool.Pool
	userRepo     repository.UserRepository
	redisClient  *redis.Client
	pollInterval time.Duration
	batchSize    int32
}

// NewDispatcher создает новый экземпляр диспетчера Outbox для auth-service
func NewDispatcher(
	pool *pgxpool.Pool,
	userRepo repository.UserRepository,
	redisClient *redis.Client,
) *Dispatcher {
	return &Dispatcher{
		pool:         pool,
		userRepo:     userRepo,
		redisClient:  redisClient,
		pollInterval: defaultPollInterval,
		batchSize:    defaultBatchSize,
	}
}

// Start запускает цикл диспетчера и фоновую очистку обработанных событий
func (d *Dispatcher) Start(ctx context.Context) error {
	logrus.Info("Starting Auth Transactional Outbox dispatcher")

	go d.startCleanupLoop(ctx)

	notifyCh := make(chan struct{}, 1)
	if d.pool != nil {
		go d.startNotifyListener(ctx, notifyCh)
	}

	interval := d.pollInterval
	if d.pool != nil {
		interval = fallbackPollInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Первоначальный прогон при старте
	d.processPendingEvents(ctx)

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping Auth Transactional Outbox dispatcher")
			return nil

		case <-notifyCh:
			d.processPendingEvents(ctx)

		case <-ticker.C:
			d.processPendingEvents(ctx)
		}
	}
}

// startNotifyListener слушает PostgreSQL LISTEN на канале auth_outbox_channel
func (d *Dispatcher) startNotifyListener(ctx context.Context, notifyCh chan<- struct{}) {
	for {
		if ctx.Err() != nil {
			return
		}

		conn, err := d.pool.Acquire(ctx)
		if err != nil {
			logrus.WithError(err).Warn("Auth Outbox: failed to acquire connection for LISTEN, retrying in 2s")
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
			logrus.WithError(err).Warn("Auth Outbox: failed to execute LISTEN, retrying in 2s")
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		logrus.Infof("Auth Outbox: successfully subscribed to %s via PostgreSQL LISTEN", outboxChannel)

		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				conn.Release()
				if ctx.Err() != nil {
					return
				}
				logrus.WithError(err).Warn("Auth Outbox: connection lost while waiting for notification, reconnecting...")
				break
			}

			if notification != nil {
				select {
				case notifyCh <- struct{}{}:
				default:
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

		events, err := d.userRepo.FetchPendingOutboxEvents(ctx, d.batchSize)
		if err != nil {
			logrus.WithError(err).Error("Auth Outbox: failed to fetch pending events")
			return
		}

		if len(events) == 0 {
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
					"event_id":     event.ID,
					"event_type":   event.EventType,
					"aggregate_id": event.AggregateID,
				}).Error("Auth Outbox: failed to process event")

				if markErr := d.userRepo.MarkOutboxEventFailed(ctx, event.ID, err.Error()); markErr != nil {
					logrus.WithError(markErr).WithField("event_id", event.ID).Error("Auth Outbox: failed to mark event as failed")
				}
			} else {
				metrics.OutboxEventsProcessedTotal.WithLabelValues(event.EventType).Inc()
				metrics.OutboxProcessingDuration.WithLabelValues(event.EventType).Observe(duration)

				if markErr := d.userRepo.MarkOutboxEventProcessed(ctx, event.ID); markErr != nil {
					logrus.WithError(markErr).WithField("event_id", event.ID).Error("Auth Outbox: failed to mark event as processed")
				}
			}
		}

		if len(events) < int(d.batchSize) {
			return
		}
	}
}

// processEvent диспетчеризирует обработку отдельного события
func (d *Dispatcher) processEvent(ctx context.Context, event sqlc.OutboxEvent) error {
	switch event.EventType {
	case "seller_renamed":
		return d.handleSellerRenamed(ctx, event)

	default:
		logrus.WithFields(logrus.Fields{
			"event_id":   event.ID,
			"event_type": event.EventType,
		}).Warn("Auth Outbox: unknown event type, skipping")
		return nil
	}
}

// handleSellerRenamed публикует событие seller_renamed в Redis Streams events:orders
func (d *Dispatcher) handleSellerRenamed(ctx context.Context, event sqlc.OutboxEvent) error {
	sellerID := event.AggregateID
	var name string

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(event.Payload, &payloadMap); err == nil {
		if n, ok := payloadMap["name"].(string); ok && n != "" {
			name = n
		}
	}

	if name == "" {
		user, err := d.userRepo.FindByID(sellerID)
		if err != nil {
			logrus.WithError(err).WithField("seller_id", sellerID).Warn("User not found during outbox seller_renamed processing")
			return nil
		}
		name = user.Name
	}

	if d.redisClient != nil {
		err := d.redisClient.XAdd(ctx, &redis.XAddArgs{
			Stream: eventsOrdersStream,
			Values: map[string]interface{}{
				"type":      "seller_renamed",
				"seller_id": fmt.Sprintf("%d", sellerID),
				"name":      name,
			},
		}).Err()
		if err != nil {
			return fmt.Errorf("failed to publish seller_renamed to Redis Streams: %w", err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"seller_id": sellerID,
		"name":      name,
	}).Info("Auth Outbox: successfully published seller_renamed to Redis Streams")

	return nil
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
			if err := d.userRepo.CleanupProcessedOutboxEvents(ctx); err != nil {
				logrus.WithError(err).Error("Auth Outbox: failed to cleanup processed outbox events")
			}
		}
	}
}
