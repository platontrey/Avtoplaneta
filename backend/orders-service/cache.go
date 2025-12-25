package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// CacheService определяет интерфейс для кеширования
type CacheService interface {
	GetOrders() ([]Order, error)
	SetOrders(orders []Order) error
	InvalidateOrders() error
}

// redisCacheService реализует CacheService с Redis
type redisCacheService struct {
	client *redis.Client
}

// NewCacheService создает новый сервис кеширования
func NewCacheService(client *redis.Client) CacheService {
	return &redisCacheService{
		client: client,
	}
}

// GetOrders получает заказы из кеша
func (c *redisCacheService) GetOrders() ([]Order, error) {
	key := "orders:active"

	val, err := c.client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		// Ключ не найден
		return nil, nil
	}
	if err != nil {
		logrus.WithError(err).Error("Failed to get orders from cache")
		return nil, err
	}

	var orders []Order
	if err := json.Unmarshal([]byte(val), &orders); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal orders from cache")
		return nil, err
	}

	logrus.Info("Orders retrieved from cache")
	return orders, nil
}

// SetOrders сохраняет заказы в кеш
func (c *redisCacheService) SetOrders(orders []Order) error {
	key := "orders:active"
	ttl := 10 * time.Minute // Кеш на 10 минут

	data, err := json.Marshal(orders)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal orders for cache")
		return err
	}

	if err := c.client.Set(context.Background(), key, data, ttl).Err(); err != nil {
		logrus.WithError(err).Error("Failed to set orders in cache")
		return err
	}

	logrus.Info("Orders cached successfully")
	return nil
}

// InvalidateOrders удаляет заказы из кеша
func (c *redisCacheService) InvalidateOrders() error {
	key := "orders:active"

	if err := c.client.Del(context.Background(), key).Err(); err != nil {
		logrus.WithError(err).Error("Failed to invalidate orders cache")
		return err
	}

	logrus.Info("Orders cache invalidated")
	return nil
}