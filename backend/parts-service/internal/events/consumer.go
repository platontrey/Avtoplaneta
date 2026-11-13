package events

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/logger"
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
			slog.WarnContext(ctx, "Redis client is nil, skipping event consumer")
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
		slog.WarnContext(ctx, "Invalid message format: missing type", slog.String("message_id", msg.ID))
		return nil
	}

	switch eventType {
	case "order_completed":
		return c.handleOrderCompleted(ctx, msg)
	case "defect_report_created":
		slog.InfoContext(ctx, "Legacy defect_report_created event acknowledged without processing (now processed synchronously)", slog.String("message_id", msg.ID))
		return nil
	case "seller_renamed":
		return c.handleSellerRenamed(ctx, msg)
	case "part_index_requested":
		return c.handlePartIndexRequested(ctx, msg)
	case "part_delete_requested":
		return c.handlePartDeleteRequested(ctx, msg)
	case "user_action":
		return c.handleUserAction(ctx, msg)
	default:
		slog.WarnContext(ctx, "Unknown event type",
			slog.String("message_id", msg.ID),
			slog.String("type", eventType))
		return nil
	}
}

// handleOrderCompleted обрабатывает событие завершения заказа
func (c *RedisEventConsumer) handleOrderCompleted(ctx context.Context, msg redis.XMessage) error {
	orderIDStr, ok := msg.Values["order_id"].(string)
	if !ok {
		slog.WarnContext(ctx, "Invalid order_completed event: missing order_id", slog.String("message_id", msg.ID))
		return nil
	}

	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid order_id format", logger.Err(err), slog.String("order_id", orderIDStr))
		return err
	}

	amountStr, ok := msg.Values["amount"].(string)
	if !ok {
		slog.WarnContext(ctx, "Invalid order_completed event: missing amount", slog.String("message_id", msg.ID))
		return nil
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid amount format", logger.Err(err), slog.String("amount", amountStr))
		return err
	}

	// Обновляем статистику доходов
	if err := c.service.UpdateEarnings(ctx, amount); err != nil {
		slog.ErrorContext(ctx, "Failed to update earnings", logger.Err(err), slog.Uint64("order_id", orderID))
		return err
	}

	slog.InfoContext(ctx, "Successfully processed order completed event",
		slog.Uint64("order_id", orderID),
		slog.Float64("amount", amount))

	return nil
}

// handleUserAction обрабатывает событие действия пользователя
func (c *RedisEventConsumer) handleUserAction(ctx context.Context, msg redis.XMessage) error {
	userID, ok := msg.Values["user_id"].(string)
	if !ok {
		slog.WarnContext(ctx, "Invalid user_action event: missing user_id", slog.String("message_id", msg.ID))
		return nil
	}

	action, ok := msg.Values["action"].(string)
	if !ok {
		slog.WarnContext(ctx, "Invalid user_action event: missing action", slog.String("message_id", msg.ID))
		return nil
	}

	slog.InfoContext(ctx, "User action logged via Redis Streams",
		slog.String("user_id", userID),
		slog.String("action", action),
		slog.Any("details", msg.Values["details"]))

	return nil
}

func (c *RedisEventConsumer) handlePartIndexRequested(ctx context.Context, msg redis.XMessage) error {
	partIDStr, ok := msg.Values["part_id"].(string)
	if !ok {
		slog.WarnContext(ctx, "Invalid part_index_requested event: missing part_id", slog.String("message_id", msg.ID))
		return nil
	}

	partID, err := strconv.ParseInt(partIDStr, 10, 64)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid part_id format in part_index_requested", logger.Err(err), slog.String("part_id", partIDStr))
		return err
	}

	part, err := c.service.GetPartByID(ctx, partID)
	if err != nil {
		// Запчасть могла быть удалена до индексации
		slog.InfoContext(ctx, "Part not found for indexing, skipping", slog.Int64("part_id", partID))
		return nil
	}

	if err := search.IndexPart(part); err != nil {
		slog.ErrorContext(ctx, "Failed to index part in Elasticsearch", logger.Err(err), slog.Int64("part_id", partID))
		return err
	}

	slog.InfoContext(ctx, "Successfully indexed part in Elasticsearch via Redis Streams", slog.Int64("part_id", partID))
	return nil
}

func (c *RedisEventConsumer) handlePartDeleteRequested(ctx context.Context, msg redis.XMessage) error {
	partIDStr, ok := msg.Values["part_id"].(string)
	if !ok {
		slog.WarnContext(ctx, "Invalid part_delete_requested event: missing part_id", slog.String("message_id", msg.ID))
		return nil
	}

	partID, err := strconv.ParseInt(partIDStr, 10, 64)
	if err != nil {
		slog.ErrorContext(ctx, "Invalid part_id format in part_delete_requested", logger.Err(err), slog.String("part_id", partIDStr))
		return err
	}

	if err := search.DeletePartFromIndex(partID); err != nil {
		slog.ErrorContext(ctx, "Failed to remove part from Elasticsearch index", logger.Err(err), slog.Int64("part_id", partID))
		return err
	}

	slog.InfoContext(ctx, "Successfully deleted part from Elasticsearch via Redis Streams", slog.Int64("part_id", partID))
	return nil
}

// handleSellerRenamed чинит копию имени продавца в запчастях.
func (c *RedisEventConsumer) handleSellerRenamed(ctx context.Context, msg redis.XMessage) error {
	sellerIDRaw, _ := msg.Values["seller_id"].(string)
	name, _ := msg.Values["name"].(string)

	sellerID, err := strconv.ParseInt(strings.TrimSpace(sellerIDRaw), 10, 64)
	if err != nil || sellerID <= 0 || strings.TrimSpace(name) == "" {
		slog.WarnContext(ctx, "Invalid seller_renamed event", slog.String("message_id", msg.ID))
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
		slog.WarnContext(ctx, "Failed to reindex parts after seller rename", logger.Err(err), slog.Int64("seller_id", sellerID))
	}

	slog.InfoContext(ctx, "Имя продавца обновлено в запчастях",
		slog.Int64("seller_id", sellerID),
		slog.String("name", name),
		slog.Int("parts", len(updated)))
	return nil
}
