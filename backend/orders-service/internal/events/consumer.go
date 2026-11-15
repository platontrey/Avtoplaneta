package events

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/redisstream"
	"orders-service/internal/repository"
)

// RedisEventConsumer слушает события из Redis Streams
type RedisEventConsumer struct {
	worker    *redisstream.ConsumerGroupWorker
	orderRepo repository.OrderRepository
	cache     repository.CacheService
}

// NewRedisEventConsumer создает новый потребитель событий
func NewRedisEventConsumer(client *redis.Client, orderRepo repository.OrderRepository, cache repository.CacheService) *RedisEventConsumer {
	consumer := &RedisEventConsumer{
		orderRepo: orderRepo,
		cache:     cache,
	}

	cfg := redisstream.ConsumerConfig{
		Stream:        constants.StreamOrdersEvents,
		Group:         "orders-service-group",
		BatchSize:     10,
		BlockDuration: 5 * time.Second,
	}

	consumer.worker = redisstream.NewConsumerGroupWorker(client, consumer.processMessage, cfg)
	return consumer
}

// Start запускает цикл прослушивания сообщений через универсальный pkg/redisstream
func (c *RedisEventConsumer) Start(ctx context.Context) error {
	return c.worker.Start(ctx)
}

func (c *RedisEventConsumer) processMessage(ctx context.Context, msg redis.XMessage) error {
	eventType, ok := msg.Values["type"].(string)
	if !ok {
		return nil
	}

	switch eventType {
	case "seller_renamed":
		return c.handleSellerRenamed(ctx, msg)
	case "part_index_requested", "part_delete_requested":
		if c.cache != nil {
			if err := c.cache.InvalidateOrders(); err != nil {
				slog.WarnContext(ctx, "Failed to invalidate orders cache after part update", logger.Err(err))
			}
		}
		return nil
	default:
		// Игнорируем события других типов (order_completed, etc.)
		return nil
	}
}

func (c *RedisEventConsumer) handleSellerRenamed(ctx context.Context, msg redis.XMessage) error {
	sellerIDRaw, _ := msg.Values["seller_id"].(string)
	name, _ := msg.Values["name"].(string)

	sellerID, err := strconv.ParseInt(strings.TrimSpace(sellerIDRaw), 10, 64)
	if err != nil || sellerID <= 0 || strings.TrimSpace(name) == "" {
		slog.WarnContext(ctx, "Invalid seller_renamed event received in orders-service", slog.String("message_id", msg.ID))
		return nil
	}

	if err := c.orderRepo.UpdateSellerName(ctx, sellerID, name); err != nil {
		slog.ErrorContext(ctx, "Failed to update seller name in orders table", logger.Err(err), slog.Int64("seller_id", sellerID), slog.String("name", name))
		return err
	}

	if c.cache != nil {
		if err := c.cache.InvalidateOrders(); err != nil {
			slog.WarnContext(ctx, "Failed to invalidate orders cache after seller rename", logger.Err(err))
		}
	}

	slog.InfoContext(ctx, "Seller name updated in orders successfully", slog.Int64("seller_id", sellerID), slog.String("name", name))
	return nil
}
