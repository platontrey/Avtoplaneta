package redisstream

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewConsumerGroupWorker_Defaults(t *testing.T) {
	handler := func(ctx context.Context, msg redis.XMessage) error {
		return nil
	}

	worker := NewConsumerGroupWorker(nil, handler, ConsumerConfig{
		Stream: "test:stream",
		Group:  "test-group",
	})

	assert.NotNil(t, worker)
	assert.Equal(t, int64(10), worker.cfg.BatchSize)
	assert.Equal(t, 5*time.Second, worker.cfg.BlockDuration)
	assert.Equal(t, 30*time.Second, worker.cfg.RecoveryPeriod)
	assert.Equal(t, int64(5), worker.cfg.MaxDeliveryAttempts)
	assert.NotEmpty(t, worker.cfg.Consumer)
}

func TestConsumerGroupWorker_NilClient_NoPanic(t *testing.T) {
	handler := func(ctx context.Context, msg redis.XMessage) error {
		return nil
	}

	worker := NewConsumerGroupWorker(nil, handler, ConsumerConfig{
		Stream: "test:stream",
		Group:  "test-group",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := worker.Start(ctx)
	assert.NoError(t, err)
}
