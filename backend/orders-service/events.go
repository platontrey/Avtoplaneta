package main

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// EventPublisher определяет интерфейс для публикации событий
type EventPublisher interface {
	PublishOrderCompleted(ctx context.Context, orderID uint, amount float64) error
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
func (p *RedisEventPublisher) PublishOrderCompleted(ctx context.Context, orderID uint, amount float64) error {
	event := map[string]interface{}{
		"type":     "order_completed",
		"order_id": orderID,
		"amount":   amount,
		"ts":       time.Now().Unix(),
	}

	err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: "events:orders",
		Values: event,
	}).Err()

	if err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to publish order completed event")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"amount":   amount,
	}).Info("Order completed event published to Redis Streams")

	return nil
}

// PublishUserAction публикует событие действия пользователя
func (p *RedisEventPublisher) PublishUserAction(ctx context.Context, userID, action string, details map[string]interface{}) error {
	event := map[string]interface{}{
		"type":    "user_action",
		"user_id": userID,
		"action":  action,
		"details": details,
		"ts":      time.Now().Unix(),
	}

	err := p.client.XAdd(ctx, &redis.XAddArgs{
		Stream: "events:orders",
		Values: event,
	}).Err()

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"action":  action,
		}).Error("Failed to publish user action event")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"action":  action,
	}).Info("User action event published to Redis Streams")

	return nil
}