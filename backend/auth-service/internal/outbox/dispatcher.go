package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"auth-service/internal/repository"
	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/metrics"
	"avtoplaneta/pkg/outbox"
)

// Dispatcher отвечает за надежную асинхронную обработку событий transactional outbox
// в auth-service на базе универсального движка pkg/outbox.Dispatcher.
type Dispatcher struct {
	userRepo    repository.UserRepository
	redisClient *redis.Client
	inner       *outbox.Dispatcher
}

type outboxStorageAdapter struct {
	repo repository.UserRepository
}

func (a *outboxStorageAdapter) FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]outbox.Event, error) {
	rows, err := a.repo.FetchPendingOutboxEvents(ctx, limit)
	if err != nil {
		return nil, err
	}
	events := make([]outbox.Event, len(rows))
	for i, r := range rows {
		events[i] = outbox.Event{
			ID:            r.ID,
			AggregateType: r.AggregateType,
			AggregateID:   r.AggregateID,
			EventType:     r.EventType,
			Payload:       r.Payload,
			RetryCount:    r.RetryCount,
		}
	}
	return events, nil
}

func (a *outboxStorageAdapter) MarkOutboxEventProcessed(ctx context.Context, id int64) error {
	return a.repo.MarkOutboxEventProcessed(ctx, id)
}

func (a *outboxStorageAdapter) MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error {
	return a.repo.MarkOutboxEventFailed(ctx, id, lastErr)
}

func (a *outboxStorageAdapter) CleanupProcessedOutboxEvents(ctx context.Context) error {
	return a.repo.CleanupProcessedOutboxEvents(ctx)
}

// NewDispatcher инициализирует диспетчер Outbox поверх pkg/outbox
func NewDispatcher(pool *pgxpool.Pool, userRepo repository.UserRepository, redisClient *redis.Client) *Dispatcher {
	d := &Dispatcher{
		userRepo:    userRepo,
		redisClient: redisClient,
	}

	storage := &outboxStorageAdapter{repo: userRepo}
	opts := outbox.Options{
		Name:         "Auth",
		Channel:      constants.OutboxChannelAuth,
		BatchSize:    50,
		PollInterval: 10 * time.Millisecond,
		Metrics:      metrics.NewOutboxMetrics("auth-service"),
	}

	d.inner = outbox.NewDispatcher(pool, storage, d.processEvent, opts)
	return d
}

// Start запускает цикл диспетчера через pkg/outbox
func (d *Dispatcher) Start(ctx context.Context) error {
	return d.inner.Start(ctx)
}

// processEvent обрабатывает событие
func (d *Dispatcher) processEvent(ctx context.Context, event outbox.Event) error {
	switch event.EventType {
	case "seller_renamed":
		return d.handleSellerRenamed(ctx, event)
	default:
		slog.WarnContext(ctx, "Auth Outbox: unknown event type, skipping",
			slog.Int64("event_id", event.ID),
			slog.String("event_type", event.EventType),
		)
		return nil
	}
}

// handleSellerRenamed публикует событие seller_renamed в Redis Streams events:orders
func (d *Dispatcher) handleSellerRenamed(ctx context.Context, event outbox.Event) error {
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
			slog.WarnContext(ctx, "User not found during outbox seller_renamed processing",
				logger.Err(err),
				slog.Int64("seller_id", sellerID),
			)
			return nil
		}
		name = user.Name
	}

	if d.redisClient != nil {
		err := d.redisClient.XAdd(ctx, &redis.XAddArgs{
			Stream: constants.StreamOrdersEvents,
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

	slog.InfoContext(ctx, "Auth Outbox: successfully published seller_renamed to Redis Streams",
		slog.Int64("seller_id", sellerID),
		slog.String("name", name),
	)

	return nil
}
