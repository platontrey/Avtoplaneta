package redisstream

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Handler функция обработки одного входящего сообщения Redis Streams
type Handler func(ctx context.Context, msg redis.XMessage) error

// ConsumerConfig конфигурация Consumer Group воркера
type ConsumerConfig struct {
	Stream              string        // Имя стрима (например, "events:orders")
	Group               string        // Имя Consumer Group (например, "parts-service-group")
	Consumer            string        // Имя консьюмера (по умолчанию: hostname-pid)
	BatchSize           int64         // Размер выборки (по умолчанию: 10)
	BlockDuration       time.Duration // Длительность блокировки XReadGroup (по умолчанию: 5s)
	RecoveryPeriod      time.Duration // Периодичность проверки зависших сообщений PEL (по умолчанию: 30s)
	MaxDeliveryAttempts int64         // Число попыток доставки до сброса в DLQ (по умолчанию: 5)
	EnableDLQ           bool          // Отправлять ли сообщения в Dead Letter Queue (<stream>:dead)
	Logger              *logrus.Entry // Опциональный логгер с контекстом сервиса
}

// ConsumerGroupWorker реализует надежную At-Least-Once обработку сообщений из Redis Streams
type ConsumerGroupWorker struct {
	client  *redis.Client
	handler Handler
	cfg     ConsumerConfig
	log     *logrus.Entry
}

// NewConsumerGroupWorker создает настроенный воркер
func NewConsumerGroupWorker(client *redis.Client, handler Handler, cfg ConsumerConfig) *ConsumerGroupWorker {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
	}
	if cfg.BlockDuration <= 0 {
		cfg.BlockDuration = 5 * time.Second
	}
	if cfg.RecoveryPeriod <= 0 {
		cfg.RecoveryPeriod = 30 * time.Second
	}
	if cfg.MaxDeliveryAttempts <= 0 {
		cfg.MaxDeliveryAttempts = 5
	}
	if cfg.Consumer == "" {
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "worker"
		}
		cfg.Consumer = fmt.Sprintf("%s-%d", hostname, os.Getpid())
	}

	logger := cfg.Logger
	if logger == nil {
		logger = logrus.WithFields(logrus.Fields{
			"component": "redis-consumer",
			"stream":    cfg.Stream,
			"group":     cfg.Group,
		})
	}

	return &ConsumerGroupWorker{
		client:  client,
		handler: handler,
		cfg:     cfg,
		log:     logger,
	}
}

// Start запускает цикл прослушивания сообщений и фоновое восстановление PEL
func (w *ConsumerGroupWorker) Start(ctx context.Context) error {
	if w.client == nil {
		w.log.Warn("Redis client is nil, skipping consumer worker")
		return nil
	}

	// 1. Создаем Consumer Group, если она еще не создана
	err := w.client.XGroupCreateMkStream(ctx, w.cfg.Stream, w.cfg.Group, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		w.log.WithError(err).Error("Failed to create consumer group")
		return fmt.Errorf("redis stream create group: %w", err)
	}

	w.log.Infof("Starting Redis Streams consumer (worker: %s)", w.cfg.Consumer)

	// 2. Фоновый цикл восстановления зависших сообщений (PEL)
	go w.startRecoveryLoop(ctx)

	// 3. Первоначальное восстановление при старте процесса
	w.recoverPendingMessages(ctx)

	// 4. Основной цикл вычитки новых сообщений (">")
	for {
		select {
		case <-ctx.Done():
			w.log.Info("Stopping Redis Streams consumer")
			return nil

		default:
			streams, err := w.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    w.cfg.Group,
				Consumer: w.cfg.Consumer,
				Streams:  []string{w.cfg.Stream, ">"},
				Count:    w.cfg.BatchSize,
				Block:    w.cfg.BlockDuration,
			}).Result()

			if err != nil && !errors.Is(err, redis.Nil) {
				w.log.WithError(err).Error("Failed to read from Redis Streams")
				time.Sleep(time.Second) // Защита от спама при временном падении Redis
				continue
			}

			if len(streams) == 0 {
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					if err := w.handler(ctx, msg); err != nil {
						w.log.WithError(err).WithField("message_id", msg.ID).Error("Failed to process stream message, leaving in PEL")
						continue
					}

					if err := w.client.XAck(ctx, w.cfg.Stream, w.cfg.Group, msg.ID).Err(); err != nil {
						w.log.WithError(err).WithField("message_id", msg.ID).Error("Failed to ACK stream message")
					}
				}
			}
		}
	}
}

// startRecoveryLoop периодически запускает обработку зависших сообщений
func (w *ConsumerGroupWorker) startRecoveryLoop(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.RecoveryPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.recoverPendingMessages(ctx)
		}
	}
}

// recoverPendingMessages читает зависшие сообщения воркера (ID "0")
func (w *ConsumerGroupWorker) recoverPendingMessages(ctx context.Context) {
	streams, err := w.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    w.cfg.Group,
		Consumer: w.cfg.Consumer,
		Streams:  []string{w.cfg.Stream, "0"},
		Count:    w.cfg.BatchSize,
		Block:    -1,
	}).Result()

	if err != nil {
		if !errors.Is(err, redis.Nil) {
			w.log.WithError(err).Error("Failed to read pending messages from Redis PEL")
		}
		return
	}

	for _, stream := range streams {
		for _, msg := range stream.Messages {
			w.log.WithField("message_id", msg.ID).Info("Recovering pending message from PEL")

			deliveryCount := w.getDeliveryCount(ctx, msg.ID)
			if deliveryCount > w.cfg.MaxDeliveryAttempts {
				w.log.WithFields(logrus.Fields{
					"message_id":       msg.ID,
					"delivery_count":   deliveryCount,
					"max_attempts":     w.cfg.MaxDeliveryAttempts,
				}).Warn("Message exceeded max delivery attempts")

				if w.cfg.EnableDLQ {
					w.sendToDLQ(ctx, msg, fmt.Errorf("exceeded max delivery attempts (%d)", deliveryCount))
				}

				// Подтверждаем в основном стриме, чтобы убрать из очереди зависших
				_ = w.client.XAck(ctx, w.cfg.Stream, w.cfg.Group, msg.ID).Err()
				continue
			}

			if err := w.handler(ctx, msg); err != nil {
				w.log.WithError(err).WithField("message_id", msg.ID).Error("Failed to process recovered message")
				continue
			}

			if err := w.client.XAck(ctx, w.cfg.Stream, w.cfg.Group, msg.ID).Err(); err != nil {
				w.log.WithError(err).WithField("message_id", msg.ID).Error("Failed to acknowledge recovered message")
			}
		}
	}
}

// getDeliveryCount возвращает число попыток доставки через XPendingExt
func (w *ConsumerGroupWorker) getDeliveryCount(ctx context.Context, messageID string) int64 {
	pending, err := w.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream: w.cfg.Stream,
		Group:  w.cfg.Group,
		Start:  messageID,
		End:    messageID,
		Count:  1,
	}).Result()

	if err != nil || len(pending) == 0 {
		return 1
	}

	return pending[0].RetryCount
}

// sendToDLQ отправляет зависшее сообщение в Dead Letter Queue (<stream>:dead)
func (w *ConsumerGroupWorker) sendToDLQ(ctx context.Context, msg redis.XMessage, procErr error) {
	dlqValues := make(map[string]interface{}, len(msg.Values)+3)
	for k, v := range msg.Values {
		dlqValues[k] = v
	}
	dlqValues["dlq_error"] = procErr.Error()
	dlqValues["dlq_time"] = time.Now().Format(time.RFC3339)
	dlqValues["original_id"] = msg.ID

	dlqStream := w.cfg.Stream + ":dead"
	err := w.client.XAdd(ctx, &redis.XAddArgs{
		Stream: dlqStream,
		Values: dlqValues,
	}).Err()

	if err != nil {
		w.log.WithError(err).WithFields(logrus.Fields{
			"original_id": msg.ID,
			"dlq_stream":  dlqStream,
		}).Error("Failed to publish message to DLQ")
	} else {
		w.log.WithFields(logrus.Fields{
			"original_id": msg.ID,
			"dlq_stream":  dlqStream,
		}).Info("Message successfully routed to DLQ")
	}
}
