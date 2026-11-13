package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/metrics"
	"avtoplaneta/pkg/outbox"
	"orders-service/internal/repository"
)

// Dispatcher отвечает за надежную асинхронную обработку событий transactional outbox
// в orders-service на базе универсального движка pkg/outbox.Dispatcher.
type Dispatcher struct {
	pool         *pgxpool.Pool
	orderRepo    repository.OrderRepository
	partRepo     repository.PartRepositoryForOrders
	redisClient  *redis.Client
	cache        repository.CacheService
	pollInterval time.Duration
	batchSize    int32
	inner        *outbox.Dispatcher
}

type outboxStorageAdapter struct {
	repo repository.OrderRepository
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
		pollInterval: 10 * time.Millisecond,
		batchSize:    50,
	}
}

// Start запускает цикл диспетчера через универсальный pkg/outbox
func (d *Dispatcher) Start(ctx context.Context) error {
	storage := &outboxStorageAdapter{repo: d.orderRepo}
	opts := outbox.Options{
		Name:         "Orders",
		Channel:      constants.OutboxChannelOrders,
		BatchSize:    d.batchSize,
		PollInterval: d.pollInterval,
		Metrics:      metrics.NewOutboxMetrics("orders-service"),
	}

	d.inner = outbox.NewDispatcher(d.pool, storage, d.processEvent, opts)
	return d.inner.Start(ctx)
}

// processEvent диспетчеризирует обработку отдельного события
func (d *Dispatcher) processEvent(ctx context.Context, event outbox.Event) error {
	switch event.EventType {
	case "order_completed":
		return d.handleOrderCompleted(ctx, event)
	case "user_action":
		return d.handleUserAction(ctx, event)
	default:
		slog.WarnContext(ctx, "Orders Outbox: unknown event type, skipping",
			slog.Int64("event_id", event.ID),
			slog.String("event_type", event.EventType))
		return nil
	}
}

// handleOrderCompleted выполняет гарантированное списание запчастей в parts-service и отправку в Redis Stream
func (d *Dispatcher) handleOrderCompleted(ctx context.Context, event outbox.Event) error {
	orderID := event.AggregateID
	order, err := d.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		slog.WarnContext(ctx, "Order not found during outbox completion processing",
			logger.Err(err),
			slog.Int64("order_id", orderID))
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
				slog.WarnContext(ctx, "Part not found in parts-service during outbox processing",
					logger.Err(errPart),
					slog.Int64("part_id", item.PartID))
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
			Stream: constants.StreamOrdersEvents,
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
			slog.WarnContext(ctx, "Orders Outbox: failed to invalidate orders cache", logger.Err(err))
		}
	}

	return nil
}

// handleUserAction публикует аудит-действие в Redis Streams
func (d *Dispatcher) handleUserAction(ctx context.Context, event outbox.Event) error {
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
		Stream: constants.StreamOrdersEvents,
		Values: values,
	}).Err()
}
