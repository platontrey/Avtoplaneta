package events

import (
	"context"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/redisstream"
	"parts-service/internal/search"
	"parts-service/internal/service"
)

// EventConsumer определяет интерфейс для обработки событий
type EventConsumer interface {
	Start(ctx context.Context) error
}

// RedisEventConsumer реализует EventConsumer с использованием Redis Streams через pkg/redisstream
type RedisEventConsumer struct {
	client   *redis.Client
	service  service.InventoryService
	group    string
	consumer string
	stream   string
	worker   *redisstream.ConsumerGroupWorker
}

// NewEventConsumer создает новый consumer событий
func NewEventConsumer(client *redis.Client, service service.InventoryService) EventConsumer {
	c := &RedisEventConsumer{
		client:   client,
		service:  service,
		group:    "parts-service-group",
		consumer: "worker-1",
		stream:   constants.StreamOrdersEvents,
	}

	if client != nil {
		c.worker = redisstream.NewConsumerGroupWorker(client, c.processMessage, redisstream.ConsumerConfig{
			Stream:    c.stream,
			Group:     c.group,
			Consumer:  c.consumer,
			EnableDLQ: true,
		})
	}

	return c
}

// Start запускает обработку событий через универсальный ConsumerGroupWorker
func (c *RedisEventConsumer) Start(ctx context.Context) error {
	if c.worker == nil {
		if c.client == nil {
			logrus.Warn("Redis client is nil, skipping event consumer")
			return nil
		}
		c.worker = redisstream.NewConsumerGroupWorker(c.client, c.processMessage, redisstream.ConsumerConfig{
			Stream:    c.stream,
			Group:     c.group,
			Consumer:  c.consumer,
			EnableDLQ: true,
		})
	}
	return c.worker.Start(ctx)
}

// processMessage обрабатывает отдельное входящее сообщение
func (c *RedisEventConsumer) processMessage(ctx context.Context, msg redis.XMessage) error {
	eventType, ok := msg.Values["type"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid message format: missing type")
		return nil
	}

	switch eventType {
	case "order_completed":
		return c.handleOrderCompleted(ctx, msg)
	case "defect_report_created":
		logrus.WithField("message_id", msg.ID).Info("Legacy defect_report_created event acknowledged without processing (now processed synchronously)")
		return nil
	case "seller_renamed":
		return c.handleSellerRenamed(ctx, msg)
	case "part_index_requested":
		return c.handlePartIndexRequested(ctx, msg)
	case "part_delete_requested":
		return c.handlePartDeleteRequested(ctx, msg)
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

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"action":  action,
		"details": msg.Values["details"],
	}).Info("User action logged via Redis Streams")

	return nil
}

func (c *RedisEventConsumer) handlePartIndexRequested(ctx context.Context, msg redis.XMessage) error {
	partIDStr, ok := msg.Values["part_id"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid part_index_requested event: missing part_id")
		return nil
	}

	partID, err := strconv.ParseInt(partIDStr, 10, 64)
	if err != nil {
		logrus.WithError(err).WithField("part_id", partIDStr).Error("Invalid part_id format in part_index_requested")
		return err
	}

	part, err := c.service.GetPartByID(ctx, partID)
	if err != nil {
		// Запчасть могла быть удалена до индексации
		logrus.WithField("part_id", partID).Info("Part not found for indexing, skipping")
		return nil
	}

	if err := search.IndexPart(part); err != nil {
		logrus.WithError(err).WithField("part_id", partID).Error("Failed to index part in Elasticsearch")
		return err
	}

	logrus.WithField("part_id", partID).Info("Successfully indexed part in Elasticsearch via Redis Streams")
	return nil
}

func (c *RedisEventConsumer) handlePartDeleteRequested(ctx context.Context, msg redis.XMessage) error {
	partIDStr, ok := msg.Values["part_id"].(string)
	if !ok {
		logrus.WithField("message_id", msg.ID).Warn("Invalid part_delete_requested event: missing part_id")
		return nil
	}

	partID, err := strconv.ParseInt(partIDStr, 10, 64)
	if err != nil {
		logrus.WithError(err).WithField("part_id", partIDStr).Error("Invalid part_id format in part_delete_requested")
		return err
	}

	if err := search.DeletePartFromIndex(partID); err != nil {
		logrus.WithError(err).WithField("part_id", partID).Error("Failed to remove part from Elasticsearch index")
		return err
	}

	logrus.WithField("part_id", partID).Info("Successfully deleted part from Elasticsearch via Redis Streams")
	return nil
}

// handleSellerRenamed чинит копию имени продавца в запчастях.
func (c *RedisEventConsumer) handleSellerRenamed(ctx context.Context, msg redis.XMessage) error {
	sellerIDRaw, _ := msg.Values["seller_id"].(string)
	name, _ := msg.Values["name"].(string)

	sellerID, err := strconv.ParseInt(strings.TrimSpace(sellerIDRaw), 10, 64)
	if err != nil || sellerID <= 0 || strings.TrimSpace(name) == "" {
		logrus.WithField("message_id", msg.ID).Warn("Invalid seller_renamed event")
		return nil
	}

	updated, err := c.service.RenameSeller(ctx, sellerID, name)
	if err != nil {
		return err
	}
	if len(updated) == 0 {
		return nil
	}

	if err := search.BulkIndexParts(ctx, updated); err != nil {
		logrus.WithError(err).WithField("seller_id", sellerID).Warn("Failed to reindex parts after seller rename")
	}

	logrus.WithFields(logrus.Fields{
		"seller_id": sellerID,
		"name":      name,
		"parts":     len(updated),
	}).Info("Имя продавца обновлено в запчастях")
	return nil
}
