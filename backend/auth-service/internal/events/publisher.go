package events

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/logger"
)

// UserEventPublisher определяет интерфейс публикации событий пользователей
type UserEventPublisher interface {
	PublishUserRenamed(ctx context.Context, userID int64, newName string) error
}

// RedisUserEventPublisher публикует события пользователей в Redis Streams
type RedisUserEventPublisher struct {
	client *redis.Client
}

// NewRedisUserEventPublisher создает новый экземпляр RedisUserEventPublisher
func NewRedisUserEventPublisher(client *redis.Client) *RedisUserEventPublisher {
	return &RedisUserEventPublisher{client: client}
}

// PublishUserRenamed отправляет событие seller_renamed в стрим events:orders
func (p *RedisUserEventPublisher) PublishUserRenamed(ctx context.Context, userID int64, newName string) error {
	if p == nil || p.client == nil {
		return nil
	}

	err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: constants.StreamOrdersEvents,
		Values: map[string]interface{}{
			"type":      "seller_renamed",
			"seller_id": fmt.Sprintf("%d", userID),
			"name":      newName,
		},
	}).Err()

	if err != nil {
		slog.ErrorContext(ctx, "Failed to publish seller_renamed event to Redis",
			logger.Err(err),
			slog.Int64("user_id", userID),
			slog.String("name", newName),
		)
		return err
	}

	slog.InfoContext(ctx, "seller_renamed event published to Redis",
		slog.Int64("user_id", userID),
		slog.String("name", newName),
	)
	return nil
}
