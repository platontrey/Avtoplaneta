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
	"orders-service/internal/domain"
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
	case "order_created":
		return d.handleOrderCreated(ctx, event)
	case "order_deleted":
		return d.handleOrderDeleted(ctx, event)
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

// handleOrderCreated выполняет гарантированное списание запчастей в parts-service при создании заказа
func (d *Dispatcher) handleOrderCreated(ctx context.Context, event outbox.Event) error {
	orderID := event.AggregateID
	order, err := d.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		slog.WarnContext(ctx, "Order not found during outbox creation processing",
			logger.Err(err),
			slog.Int64("order_id", orderID))
		return nil
	}

	return d.deductOrderItems(ctx, orderID, order.Items)
}

// handleOrderDeleted выполняет гарантированный возврат запчастей в parts-service при удалении заказа
func (d *Dispatcher) handleOrderDeleted(ctx context.Context, event outbox.Event) error {
	if d.partRepo == nil {
		return nil
	}

	var payload struct {
		OrderID int64 `json:"order_id"`
		Items   []struct {
			PartID   int64 `json:"part_id"`
			Quantity int   `json:"quantity"`
		} `json:"items"`
	}

	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		slog.WarnContext(ctx, "Failed to unmarshal order_deleted payload",
			logger.Err(err),
			slog.Int64("event_id", event.ID))
		return nil
	}

	for _, item := range payload.Items {
		if item.PartID <= 0 {
			continue
		}
		opID := fmt.Sprintf("delete-order-%d-part-%d", payload.OrderID, item.PartID)
		if err := d.partRepo.IncreaseQuantity(ctx, item.PartID, item.Quantity, opID); err != nil {
			return fmt.Errorf("failed to restore part %d quantity for order %d: %w", item.PartID, payload.OrderID, err)
		}
	}
	return nil
}

// handleOrderCompleted выполняет публикацию события в Redis Stream (остатки уже списаны при создании заказа)
func (d *Dispatcher) handleOrderCompleted(ctx context.Context, event outbox.Event) error {
	orderID := event.AggregateID
	order, err := d.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		slog.WarnContext(ctx, "Order not found during outbox completion processing",
			logger.Err(err),
			slog.Int64("order_id", orderID))
		return nil
	}

	// 1. Публикация события order_completed в Redis Stream events:orders
	if err := d.publishOrderCompleted(ctx, orderID, order, event.Payload); err != nil {
		return err
	}

	// 2. Инвалидация кэша заказов
	if d.cache != nil {
		if err := d.cache.InvalidateOrders(); err != nil {
			slog.WarnContext(ctx, "Orders Outbox: failed to invalidate orders cache", logger.Err(err))
		}
	}

	return nil
}

func (d *Dispatcher) deductOrderItems(ctx context.Context, orderID int64, items []domain.OrderItem) error {
	if d.partRepo == nil {
		return nil
	}

	for _, item := range items {
		if err := d.deductSingleItem(ctx, orderID, item); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) deductSingleItem(ctx context.Context, orderID int64, item domain.OrderItem) error {
	if item.PartID <= 0 {
		return nil
	}

	opID := fmt.Sprintf("create-order-%d-part-%d", orderID, item.PartID)
	if err := d.partRepo.DecreaseQuantity(ctx, item.PartID, item.Quantity, opID); err != nil {
		return fmt.Errorf("failed to decrease part %d quantity: %w", item.PartID, err)
	}
	return nil
}

func (d *Dispatcher) publishOrderCompleted(ctx context.Context, orderID int64, order *domain.Order, payload []byte) error {
	if d.redisClient == nil {
		return nil
	}

	totalAmount := resolveTotalAmount(order, payload)
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
	return nil
}

func resolveTotalAmount(order *domain.Order, payload []byte) float64 {
	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payload, &payloadMap); err == nil {
		if amt, ok := payloadMap["amount"].(float64); ok {
			return amt
		}
	}

	var subtotal float64
	for _, item := range order.Items {
		subtotal += item.Price * float64(item.Quantity)
	}
	return math.Max(subtotal-order.Discount, 0)
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
