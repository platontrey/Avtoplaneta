package repository

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/logger"
	"orders-service/internal/domain"
)

// CacheService определяет интерфейс для кеширования активных заказов
type CacheService interface {
	GetOrders() ([]domain.Order, error)
	SetOrders(orders []domain.Order) error
	InvalidateOrders() error
}

// redisCacheService реализует CacheService с Redis
type redisCacheService struct {
	client *redis.Client
}

// NewCacheService создает новый сервис кеширования заказов
func NewCacheService(client *redis.Client) CacheService {
	return &redisCacheService{
		client: client,
	}
}

// GetOrders получает заказы из кеша
func (c *redisCacheService) GetOrders() ([]domain.Order, error) {
	if c.client == nil {
		return nil, nil
	}
	key := "orders:active"

	val, err := c.client.Get(context.Background(), key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		slog.Error("Failed to get orders from cache", logger.Err(err))
		return nil, err
	}

	var orders []domain.Order
	if err := json.Unmarshal([]byte(val), &orders); err != nil {
		slog.Error("Failed to unmarshal orders from cache", logger.Err(err))
		return nil, err
	}

	slog.Info("Orders retrieved from cache")
	return orders, nil
}

// SetOrders сохраняет заказы в кеш на 10 минут
func (c *redisCacheService) SetOrders(orders []domain.Order) error {
	if c.client == nil {
		return nil
	}
	key := "orders:active"
	ttl := 10 * time.Minute

	data, err := json.Marshal(orders)
	if err != nil {
		slog.Error("Failed to marshal orders for cache", logger.Err(err))
		return err
	}

	if err := c.client.Set(context.Background(), key, data, ttl).Err(); err != nil {
		slog.Error("Failed to set orders in cache", logger.Err(err))
		return err
	}

	slog.Info("Orders cached successfully")
	return nil
}

// InvalidateOrders удаляет заказы из кеша
func (c *redisCacheService) InvalidateOrders() error {
	if c.client == nil {
		return nil
	}
	key := "orders:active"

	if err := c.client.Del(context.Background(), key).Err(); err != nil {
		slog.Error("Failed to invalidate orders cache", logger.Err(err))
		return err
	}

	slog.Info("Orders cache invalidated")
	return nil
}
