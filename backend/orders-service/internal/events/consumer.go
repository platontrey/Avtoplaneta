package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/logger"
	"orders-service/internal/repository"
)

// RedisEventConsumer слушает события из Redis Streams
type RedisEventConsumer struct {
	client    *redis.Client
	orderRepo repository.OrderRepository
	cache     repository.CacheService
	group     string
	consumer  string
	stream    string
}

// NewRedisEventConsumer создает новый потребитель событий
func NewRedisEventConsumer(client *redis.Client, orderRepo repository.OrderRepository, cache repository.CacheService) *RedisEventConsumer {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "orders-worker"
	}

	return &RedisEventConsumer{
		client:    client,
		orderRepo: orderRepo,
		cache:     cache,
		group:     "orders-service-group",
		consumer:  fmt.Sprintf("%s-%d", hostname, os.Getpid()),
		stream:    constants.StreamOrdersEvents,
	}
}

// Start запускает цикл прослушивания сообщений из Redis Streams
func (c *RedisEventConsumer) Start(ctx context.Context) error {
	if c.client == nil {
		slog.WarnContext(ctx, "Redis client is nil, skipping event consumer")
		return nil
	}

	err := c.client.XGroupCreateMkStream(ctx, c.stream, c.group, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		slog.ErrorContext(ctx, "Failed to create consumer group for orders-service", logger.Err(err))
		return err
	}

	slog.InfoContext(ctx, "Starting Redis Streams consumer for orders-service")

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Stopping orders-service Redis Streams consumer")
			return nil
		default:
			streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    c.group,
				Consumer: c.consumer,
				Streams:  []string{c.stream, ">"},
				Count:    10,
				Block:    time.Second * 5,
			}).Result()

			if err != nil && !errors.Is(err, redis.Nil) {
				slog.ErrorContext(ctx, "Failed to read from Redis Streams in orders-service", logger.Err(err))
				time.Sleep(time.Second)
				continue
			}

			if len(streams) == 0 {
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					if err := c.processMessage(ctx, msg); err != nil {
						slog.ErrorContext(ctx, "Failed to process message in orders-service", logger.Err(err), slog.String("message_id", msg.ID))
						continue
					}

					if err := c.client.XAck(ctx, c.stream, c.group, msg.ID).Err(); err != nil {
						slog.ErrorContext(ctx, "Failed to ACK message in orders-service", logger.Err(err), slog.String("message_id", msg.ID))
					}
				}
			}
		}
	}
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
