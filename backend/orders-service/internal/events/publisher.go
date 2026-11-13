package events

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/logger"
)

// EventPublisher определяет интерфейс для публикации событий
type EventPublisher interface {
	PublishOrderCompleted(ctx context.Context, orderID int64, amount float64) error
	PublishUserAction(ctx context.Context, userID, action string, details map[string]interface{}) error
}

// RedisEventPublisher реализует EventPublisher с использованием Redis Streams
type RedisEventPublisher struct {
	client *redis.Client
}

// NewEventPublisher создает новый publisher событий
func NewEventPublisher(client *redis.Client) EventPublisher {
	return &RedisEventPublisher{
		client: client,
	}
}

// PublishOrderCompleted публикует событие завершения заказа
func (p *RedisEventPublisher) PublishOrderCompleted(ctx context.Context, orderID int64, amount float64) error {
	if p.client == nil {
		return nil
	}

	event := map[string]interface{}{
		"type":     "order_completed",
		"order_id": orderID,
		"amount":   amount,
		"ts":       time.Now().Unix(),
	}

	err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: constants.StreamOrdersEvents,
		Values: event,
	}).Err()

	if err != nil {
		slog.ErrorContext(ctx, "Failed to publish order completed event",
			slog.Int64("order_id", orderID),
			logger.Err(err),
		)
		return err
	}

	slog.InfoContext(ctx, "Order completed event published to Redis Streams",
		slog.Int64("order_id", orderID),
		slog.Float64("amount", amount),
	)

	return nil
}

// PublishUserAction публикует событие действия пользователя
func (p *RedisEventPublisher) PublishUserAction(ctx context.Context, userID, action string, details map[string]interface{}) error {
	if p.client == nil {
		return nil
	}

	event := map[string]interface{}{
		"type":    "user_action",
		"user_id": userID,
		"action":  action,
		"details": details,
		"ts":      time.Now().Unix(),
	}

	err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: constants.StreamOrdersEvents,
		Values: event,
	}).Err()

	if err != nil {
		slog.ErrorContext(ctx, "Failed to publish user action event",
			slog.String("user_id", userID),
			slog.String("action", action),
			logger.Err(err),
		)
		return err
	}

	slog.InfoContext(ctx, "User action event published to Redis Streams",
		slog.String("user_id", userID),
		slog.String("action", action),
	)

	return nil
}
