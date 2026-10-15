package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"orders-service/db/sqlc"
	"orders-service/internal/metrics"
	"orders-service/internal/repository"
)

const (
	defaultPollInterval  = 100 * time.Millisecond
	fallbackPollInterval = 3 * time.Second
	defaultBatchSize     = 50
	cleanupInterval      = 1 * time.Hour
	outboxChannel        = "orders_outbox_channel"
	eventsOrdersStream   = "events:orders"
)

// Dispatcher отвечает за надежную асинхронную обработку событий из transactional outbox таблицы
// в orders-service и синхронизацию состояния с parts-service (через gRPC) и Redis Streams.
type Dispatcher struct {
	pool         *pgxpool.Pool
	orderRepo    repository.OrderRepository
	partRepo     repository.PartRepositoryForOrders
	redisClient  *redis.Client
	cache        repository.CacheService
	pollInterval time.Duration
	batchSize    int32
}

// NewDispatcher создает новый экземпляр диспетчера Outbox для orders-service
func NewDispatcher(
	pool *pgxpool.Pool,
	orderRepo repository.OrderRepository,
	partRepo repository.PartRepositoryForOrders,
	redisClient *redis.Client,
	cache repository.CacheService,
) *Dispatcher {
	return &Dispatcher{
		pool:         pool,
		orderRepo:    orderRepo,
		partRepo:     partRepo,
		redisClient:  redisClient,
		cache:        cache,
		pollInterval: defaultPollInterval,
		batchSize:    defaultBatchSize,
	}
}

// Start запускает цикл диспетчера и фоновую очистку обработанных событий
func (d *Dispatcher) Start(ctx context.Context) error {
	logrus.Info("Starting Orders Transactional Outbox dispatcher")

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
			logrus.Info("Stopping Orders Transactional Outbox dispatcher")
			return nil

		case <-notifyCh:
			d.processPendingEvents(ctx)

		case <-ticker.C:
			d.processPendingEvents(ctx)
		}
	}
}

// startNotifyListener слушает PostgreSQL LISTEN на канале orders_outbox_channel
func (d *Dispatcher) startNotifyListener(ctx context.Context, notifyCh chan<- struct{}) {
	for {
		if ctx.Err() != nil {
			return
		}

		conn, err := d.pool.Acquire(ctx)
		if err != nil {
			logrus.WithError(err).Warn("Orders Outbox: failed to acquire connection for LISTEN, retrying in 2s")
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
			logrus.WithError(err).Warn("Orders Outbox: failed to execute LISTEN, retrying in 2s")
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		logrus.Infof("Orders Outbox: successfully subscribed to %s via PostgreSQL LISTEN", outboxChannel)

		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				conn.Release()
				if ctx.Err() != nil {
					return
				}
				logrus.WithError(err).Warn("Orders Outbox: connection lost while waiting for notification, reconnecting...")
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

		events, err := d.orderRepo.FetchPendingOutboxEvents(ctx, d.batchSize)
		if err != nil {
			logrus.WithError(err).Error("Orders Outbox: failed to fetch pending events")
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
				}).Error("Orders Outbox: failed to process event")

				if markErr := d.orderRepo.MarkOutboxEventFailed(ctx, event.ID, err.Error()); markErr != nil {
					logrus.WithError(markErr).WithField("event_id", event.ID).Error("Orders Outbox: failed to mark event as failed")
				}
			} else {
				metrics.OutboxEventsProcessedTotal.WithLabelValues(event.EventType).Inc()
				metrics.OutboxProcessingDuration.WithLabelValues(event.EventType).Observe(duration)

				if markErr := d.orderRepo.MarkOutboxEventProcessed(ctx, event.ID); markErr != nil {
					logrus.WithError(markErr).WithField("event_id", event.ID).Error("Orders Outbox: failed to mark event as processed")
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
	case "order_completed":
		return d.handleOrderCompleted(ctx, event)

	case "user_action":
		return d.handleUserAction(ctx, event)

	default:
		logrus.WithFields(logrus.Fields{
			"event_id":   event.ID,
			"event_type": event.EventType,
		}).Warn("Orders Outbox: unknown event type, skipping")
		return nil
	}
}

// handleOrderCompleted выполняет гарантированное списание запчастей в parts-service и отправку в Redis Stream
func (d *Dispatcher) handleOrderCompleted(ctx context.Context, event sqlc.OutboxEvent) error {
	orderID := event.AggregateID
	order, err := d.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Warn("Order not found during outbox completion processing")
		return nil
	}

	// 1. Гарантированное списание запчастей через gRPC с идемпотентным operationID
	if d.partRepo != nil {
		for _, item := range order.Items {
			if item.PartID <= 0 {
				continue
			}

			part, errPart := d.partRepo.FindByID(ctx, item.PartID)
			if errPart != nil {
				logrus.WithError(errPart).WithField("part_id", item.PartID).Warn("Part not found in parts-service during outbox processing")
				continue
			}

			if part.Quantity > item.Quantity {
				opID := fmt.Sprintf("complete-order-%d-part-%d", orderID, item.PartID)
				if err := d.partRepo.DecreaseQuantity(ctx, item.PartID, item.Quantity, opID); err != nil {
					return fmt.Errorf("failed to decrease part %d quantity: %w", item.PartID, err)
				}
			} else {
				if err := d.partRepo.DeletePart(ctx, item.PartID); err != nil {
					return fmt.Errorf("failed to delete depleted part %d: %w", item.PartID, err)
				}
			}
		}
	}

	// 2. Публикация события order_completed в Redis Stream events:orders
	var totalAmount float64
	var payloadMap map[string]interface{}
	if err := json.Unmarshal(event.Payload, &payloadMap); err == nil {
		if amt, ok := payloadMap["amount"].(float64); ok {
			totalAmount = amt
		}
	}
	if totalAmount == 0 {
		var subtotal float64
		for _, item := range order.Items {
			subtotal += item.Price * float64(item.Quantity)
		}
		totalAmount = math.Max(subtotal-order.Discount, 0)
	}

	if d.redisClient != nil {
		err := d.redisClient.XAdd(ctx, &redis.XAddArgs{
			Stream: eventsOrdersStream,
			Values: map[string]interface{}{
				"type":     "order_completed",
				"order_id": fmt.Sprintf("%d", orderID),
				"amount":   fmt.Sprintf("%.2f", totalAmount),
				"ts":       time.Now().Unix(),
			},
		}).Err()
		if err != nil {
			return fmt.Errorf("failed to publish order_completed to Redis Streams: %w", err)
		}
	}

	// 3. Инвалидация кэша заказов
	if d.cache != nil {
		if err := d.cache.InvalidateOrders(); err != nil {
			logrus.WithError(err).Warn("Orders Outbox: failed to invalidate orders cache")
		}
	}

	return nil
}

// handleUserAction публикует аудит-действие в Redis Streams
func (d *Dispatcher) handleUserAction(ctx context.Context, event sqlc.OutboxEvent) error {
	if d.redisClient == nil {
		return nil
	}

	var values map[string]interface{}
	if err := json.Unmarshal(event.Payload, &values); err != nil {
		values = make(map[string]interface{})
	}
	values["type"] = "user_action"
	values["ts"] = time.Now().Unix()

	return d.redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: eventsOrdersStream,
		Values: values,
	}).Err()
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
			if err := d.orderRepo.CleanupProcessedOutboxEvents(ctx); err != nil {
				logrus.WithError(err).Error("Orders Outbox: failed to cleanup processed outbox events")
			}
		}
	}
}
