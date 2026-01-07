package main

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// EventConsumer определяет интерфейс для обработки событий
type EventConsumer interface {
	Start(ctx context.Context) error
}

// RedisEventConsumer реализует EventConsumer с использованием Redis Streams
type RedisEventConsumer struct {
	client   *redis.Client
	service  InventoryService
	group    string
	consumer string
	stream   string
}

// NewEventConsumer создает новый consumer событий
func NewEventConsumer(client *redis.Client, service InventoryService) EventConsumer {
	return &RedisEventConsumer{
		client:   client,
		service:  service,
		group:    "parts-service-group",
		consumer: "worker-1",
		stream:   "events:orders",
	}
}

// Start запускает обработку событий
func (c *RedisEventConsumer) Start(ctx context.Context) error {
	// Создаем consumer group, если не существует
	err := c.client.XGroupCreateMkStream(ctx, c.stream, c.group, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		logrus.WithError(err).Error("Failed to create consumer group")
		return err
	}

	logrus.Info("Starting Redis Streams consumer for order events")

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping Redis Streams consumer")
			return nil
		default:
			// Читаем сообщения из потока
			streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    c.group,
				Consumer: c.consumer,
				Streams:  []string{c.stream, ">"},
				Count:    10,
				Block:    time.Second * 5, // Блокируем на 5 секунд
			}).Result()

			if err != nil && !errors.Is(err, redis.Nil) {
				logrus.WithError(err).Error("Failed to read from Redis Streams")
				continue
			}

			if len(streams) == 0 {
				continue
			}

			// Обрабатываем сообщения
			for _, stream := range streams {
				for _, msg := range stream.Messages {
					if err := c.processMessage(ctx, msg); err != nil {
						logrus.WithError(err).WithField("message_id", msg.ID).Error("Failed to process message")
						continue
					}

					// Подтверждаем обработку
					if err := c.client.XAck(ctx, c.stream, c.group, msg.ID).Err(); err != nil {
						logrus.WithError(err).WithField("message_id", msg.ID).Error("Failed to acknowledge message")
					}
				}
			}
		}
	}
}

// processMessage обрабатывает отдельное сообщение
func (c *RedisEventConsumer) processMessage(ctx context.Context, msg redis.XMessage) error {
	eventType, ok := msg.Values["type"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid message format: missing type")
		return nil
	}

	switch eventType {
	case "order_completed":
		return c.handleOrderCompleted(ctx, msg)
	case "user_action":
		return c.handleUserAction(msg)
	default:
		logrus.WithFields(logrus.Fields{
			"message_id": msg.ID,
			"type":       eventType,
		}).Warn("Unknown event type")
		return nil
	}
}

// handleOrderCompleted обрабатывает событие завершения заказа
func (c *RedisEventConsumer) handleOrderCompleted(ctx context.Context, msg redis.XMessage) error {
	orderIDStr, ok := msg.Values["order_id"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid order_completed event: missing order_id")
		return nil
	}

	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		logrus.WithError(err).WithField("order_id", orderIDStr).Error("Invalid order_id format")
		return err
	}

	amountStr, ok := msg.Values["amount"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid order_completed event: missing amount")
		return nil
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		logrus.WithError(err).WithField("amount", amountStr).Error("Invalid amount format")
		return err
	}

	// Обновляем статистику доходов
	if err := c.service.UpdateEarnings(ctx, amount); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to update earnings")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"amount":   amount,
	}).Info("Successfully processed order completed event")

	return nil
}

// handleUserAction обрабатывает событие действия пользователя
func (c *RedisEventConsumer) handleUserAction(msg redis.XMessage) error {
	userID, ok := msg.Values["user_id"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid user_action event: missing user_id")
		return nil
	}

	action, ok := msg.Values["action"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid user_action event: missing action")
		return nil
	}

	// Здесь можно добавить логику для обработки действия пользователя
	// Например, отправить в messaging-service или просто залогировать
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"action":  action,
		"details": msg.Values["details"],
	}).Info("User action logged via Redis Streams")

	return nil
}
