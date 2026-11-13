package redisclient

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// New создает клиент Redis с автоматической OpenTelemetry трассировкой и проверяет подключение через Ping
func New(addr, password string) (*redis.Client, error) {
	return NewWithOptions(&redis.Options{
		Addr:     addr,
		Password: password,
	})
}

// NewWithOptions создает клиент Redis по переданным Options с автоматической OpenTelemetry трассировкой
func NewWithOptions(opts *redis.Options) (*redis.Client, error) {
	client := redis.NewClient(opts)
	Instrument(client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", opts.Addr, err)
	}

	slog.Info("Redis connected successfully", slog.String("addr", opts.Addr))
	return client, nil
}
