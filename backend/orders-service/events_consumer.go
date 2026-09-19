package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RedisEventConsumer struct {
	client    *redis.Client
	orderRepo OrderRepository
	cache     CacheService
	group     string
	consumer  string
	stream    string
}

func NewRedisEventConsumer(client *redis.Client, orderRepo OrderRepository, cache CacheService) *RedisEventConsumer {
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
		stream:    "events:orders",
	}
}

func (c *RedisEventConsumer) Start(ctx context.Context) error {
	if c.client == nil {
		logrus.Warn("Redis client is nil, skipping event consumer")
		return nil
	}

	err := c.client.XGroupCreateMkStream(ctx, c.stream, c.group, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		logrus.WithError(err).Error("Failed to create consumer group for orders-service")
		return err
	}

	logrus.Info("Starting Redis Streams consumer for orders-service")

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping orders-service Redis Streams consumer")
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
				logrus.WithError(err).Error("Failed to read from Redis Streams in orders-service")
				time.Sleep(time.Second)
				continue
			}

			if len(streams) == 0 {
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					if err := c.processMessage(ctx, msg); err != nil {
						logrus.WithError(err).WithField("message_id", msg.ID).Error("Failed to process message in orders-service")
						continue
					}

					if err := c.client.XAck(ctx, c.stream, c.group, msg.ID).Err(); err != nil {
						logrus.WithError(err).WithField("message_id", msg.ID).Error("Failed to ACK message in orders-service")
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
		logrus.WithField("message_id", msg.ID).Warn("Invalid seller_renamed event received in orders-service")
		return nil
	}

	if err := c.orderRepo.UpdateSellerName(ctx, sellerID, name); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"seller_id": sellerID,
			"name":      name,
		}).Error("Failed to update seller name in orders table")
		return err
	}

	if c.cache != nil {
		if err := c.cache.InvalidateOrders(); err != nil {
			logrus.WithError(err).Warn("Failed to invalidate orders cache after seller rename")
		}
	}

	logrus.WithFields(logrus.Fields{
		"seller_id": sellerID,
		"name":      name,
	}).Info("Seller name updated in orders successfully")
	return nil
}
