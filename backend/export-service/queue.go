package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	exportStreamKey = "stream:export_jobs"
	exportGroupKey  = "export-service-group"
	exportStatusKey = "export:status:last"
)

// ExportTask описывает задачу на экспорт каталога
type ExportTask struct {
	TaskID      string    `json:"task_id"`
	Target      string    `json:"target"` // "drom", "xml"
	TriggeredBy string    `json:"triggered_by,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ExportStatus описывает текущий или последний статус выполнения экспорта
type ExportStatus struct {
	TaskID      string     `json:"task_id"`
	Status      string     `json:"status"` // "idle", "queued", "running", "success", "failed"
	TriggeredBy string     `json:"triggered_by,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	PartsCount  int        `json:"parts_count,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// ExportQueue управляет очередью задач через Redis Streams с плавным fallback на in-process воркер
type ExportQueue struct {
	redisClient *redis.Client
	builder     *PriceListBuilder
	config      *Config

	// Локальное состояние (fallback и защита от конкурентных запусков)
	running    atomic.Bool
	statusMu   sync.RWMutex
	lastStatus ExportStatus
	localQueue chan ExportTask
}

func NewExportQueue(redisClient *redis.Client, builder *PriceListBuilder, config *Config) *ExportQueue {
	return &ExportQueue{
		redisClient: redisClient,
		builder:     builder,
		config:      config,
		localQueue:  make(chan ExportTask, 10),
		lastStatus: ExportStatus{
			Status: "idle",
		},
	}
}

// IsBusy проверяет, выполняется ли задача прямо сейчас
func (q *ExportQueue) IsBusy() bool {
	if q.running.Load() {
		return true
	}

	if q.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		status := q.GetStatus(ctx)
		// Если статус queued или running, и задача запущена менее 30 минут назад (защита от падений)
		if (status.Status == "running" || status.Status == "queued") && status.StartedAt != nil {
			if time.Since(*status.StartedAt) < 30*time.Minute {
				return true
			}
		}
	}

	return false
}

// GetStatus возвращает статус последней/текущей задачи
func (q *ExportQueue) GetStatus(ctx context.Context) ExportStatus {
	if q.redisClient != nil {
		val, err := q.redisClient.Get(ctx, exportStatusKey).Result()
		if err == nil {
			var status ExportStatus
			if jsonErr := json.Unmarshal([]byte(val), &status); jsonErr == nil {
				return status
			}
		}
	}

	q.statusMu.RLock()
	defer q.statusMu.RUnlock()
	return q.lastStatus
}

func (q *ExportQueue) setStatus(ctx context.Context, status ExportStatus) {
	q.statusMu.Lock()
	q.lastStatus = status
	q.statusMu.Unlock()

	if q.redisClient != nil {
		data, err := json.Marshal(status)
		if err == nil {
			_ = q.redisClient.Set(ctx, exportStatusKey, data, 7*24*time.Hour).Err()
		}
	}
}

// Enqueue ставит задачу на выгрузку в очередь
func (q *ExportQueue) Enqueue(ctx context.Context, target string, triggeredBy string) (ExportTask, error) {
	if q.IsBusy() {
		return ExportTask{}, errors.New("экспорт уже выполняется в данный момент")
	}

	if !q.running.CompareAndSwap(false, true) {
		return ExportTask{}, errors.New("экспорт уже выполняется в данный момент")
	}

	task := ExportTask{
		TaskID:      fmt.Sprintf("export-%d", time.Now().UnixNano()),
		Target:      target,
		TriggeredBy: triggeredBy,
		CreatedAt:   time.Now(),
	}

	now := time.Now()
	q.setStatus(ctx, ExportStatus{
		TaskID:      task.TaskID,
		Status:      "queued",
		TriggeredBy: triggeredBy,
		StartedAt:   &now,
	})

	if q.redisClient != nil {
		data, err := json.Marshal(task)
		if err == nil {
			err = q.redisClient.XAdd(ctx, &redis.XAddArgs{
				Stream: exportStreamKey,
				Values: map[string]interface{}{
					"task_id": task.TaskID,
					"payload": string(data),
				},
			}).Err()
			if err == nil {
				logrus.WithField("task_id", task.TaskID).Info("Задача экспорта опубликована в Redis Streams")
				return task, nil
			}
			logrus.WithError(err).Warn("Не удалось отправить задачу в Redis Streams, переключение на локальную очередь")
		}
	}

	// Fallback на локальную очередь
	select {
	case q.localQueue <- task:
		logrus.WithField("task_id", task.TaskID).Info("Задача экспорта поставлена в локальную очередь")
		return task, nil
	default:
		q.running.Store(false)
		return ExportTask{}, errors.New("очередь задач переполнена")
	}
}

// StartWorker запускает обработчик задач (Redis Streams и резервный локальный воркер)
func (q *ExportQueue) StartWorker(ctx context.Context) {
	go q.localWorker(ctx)

	if q.redisClient != nil {
		err := q.redisClient.XGroupCreateMkStream(ctx, exportStreamKey, exportGroupKey, "$").Err()
		if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			logrus.WithError(err).Warn("Не удалось создать Consumer Group в Redis Streams")
		} else {
			go q.redisStreamsWorker(ctx)
		}
	}
}

func (q *ExportQueue) localWorker(ctx context.Context) {
	logrus.Info("Запущен локальный in-process воркер экспорта")
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-q.localQueue:
			q.executeTask(ctx, task)
		}
	}
}

func (q *ExportQueue) redisStreamsWorker(ctx context.Context) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "export-worker"
	}
	consumerName := fmt.Sprintf("%s-%d", hostname, os.Getpid())

	logrus.WithFields(logrus.Fields{
		"stream":   exportStreamKey,
		"group":    exportGroupKey,
		"consumer": consumerName,
	}).Info("Запущен воркер Redis Streams для export-service")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			streams, err := q.redisClient.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    exportGroupKey,
				Consumer: consumerName,
				Streams:  []string{exportStreamKey, ">"},
				Count:    1,
				Block:    5 * time.Second,
			}).Result()

			if err != nil && !errors.Is(err, redis.Nil) {
				time.Sleep(time.Second)
				continue
			}

			if len(streams) == 0 {
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					payload, ok := msg.Values["payload"].(string)
					if !ok {
						_ = q.redisClient.XAck(ctx, exportStreamKey, exportGroupKey, msg.ID).Err()
						continue
					}

					var task ExportTask
					if err := json.Unmarshal([]byte(payload), &task); err != nil {
						_ = q.redisClient.XAck(ctx, exportStreamKey, exportGroupKey, msg.ID).Err()
						continue
					}

					q.executeTask(ctx, task)
					_ = q.redisClient.XAck(ctx, exportStreamKey, exportGroupKey, msg.ID).Err()
				}
			}
		}
	}
}

func (q *ExportQueue) executeTask(ctx context.Context, task ExportTask) {
	q.running.Store(true)
	defer q.running.Store(false)

	startedAt := time.Now()
	q.setStatus(ctx, ExportStatus{
		TaskID:      task.TaskID,
		Status:      "running",
		TriggeredBy: task.TriggeredBy,
		StartedAt:   &startedAt,
	})

	logrus.WithField("task_id", task.TaskID).Info("Начало выполнения задачи экспорта на Drom")

	// Сборка XML
	xmlData, count, err := q.builder.Generate(ctx)
	if err != nil {
		finishedAt := time.Now()
		errMsg := fmt.Sprintf("ошибка сборки XML: %v", err)
		logrus.WithError(err).WithField("task_id", task.TaskID).Error("Сборка XML завершилась неудачей")
		q.setStatus(ctx, ExportStatus{
			TaskID:      task.TaskID,
			Status:      "failed",
			TriggeredBy: task.TriggeredBy,
			StartedAt:   &startedAt,
			FinishedAt:  &finishedAt,
			Error:       errMsg,
		})
		return
	}

	// Отправка в Drom API
	if err := sendToDromAPI(ctx, q.config, xmlData); err != nil {
		finishedAt := time.Now()
		errMsg := fmt.Sprintf("ошибка отправки в Drom API: %v", err)
		logrus.WithError(err).WithField("task_id", task.TaskID).Error("Отправка на Drom завершилась неудачей")
		q.setStatus(ctx, ExportStatus{
			TaskID:      task.TaskID,
			Status:      "failed",
			TriggeredBy: task.TriggeredBy,
			StartedAt:   &startedAt,
			FinishedAt:  &finishedAt,
			PartsCount:  count,
			Error:       errMsg,
		})
		return
	}

	finishedAt := time.Now()
	logrus.WithFields(logrus.Fields{
		"task_id":     task.TaskID,
		"parts_count": count,
		"duration":    time.Since(startedAt).String(),
	}).Info("Экспорт на Drom успешно завершен")

	q.setStatus(ctx, ExportStatus{
		TaskID:      task.TaskID,
		Status:      "success",
		TriggeredBy: task.TriggeredBy,
		StartedAt:   &startedAt,
		FinishedAt:  &finishedAt,
		PartsCount:  count,
	})
}
