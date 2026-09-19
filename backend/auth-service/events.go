package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
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
		Stream: "events:orders",
		Values: map[string]interface{}{
			"type":      "seller_renamed",
			"seller_id": fmt.Sprintf("%d", userID),
			"name":      newName,
		},
	}).Err()

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"name":    newName,
		}).Error("Failed to publish seller_renamed event to Redis")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"name":    newName,
	}).Info("seller_renamed event published to Redis")
	return nil
}
