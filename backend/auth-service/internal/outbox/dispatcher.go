package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"auth-service/internal/repository"
	"avtoplaneta/pkg/constants"
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

// NewDispatcher создает новый экземпляр диспетчера Outbox для auth-service
func NewDispatcher(
	pool *pgxpool.Pool,
	userRepo repository.UserRepository,
	redisClient *redis.Client,
) *Dispatcher {
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
		logrus.WithFields(logrus.Fields{
			"event_id":   event.ID,
			"event_type": event.EventType,
		}).Warn("Auth Outbox: unknown event type, skipping")
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
			logrus.WithError(err).WithField("seller_id", sellerID).Warn("User not found during outbox seller_renamed processing")
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

	logrus.WithFields(logrus.Fields{
		"seller_id": sellerID,
		"name":      name,
	}).Info("Auth Outbox: successfully published seller_renamed to Redis Streams")

	return nil
}
