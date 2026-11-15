package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/redisstream"
	"export-service/internal/config"
	"export-service/internal/domain"
	"export-service/internal/service"
)

const (
	exportStreamKey = "stream:export_jobs"
	exportGroupKey  = "export-service-group"
	exportStatusKey = "export:status:last"
)

// ExportQueue управляет асинхронной очередью задач выгрузки каталога через Redis Streams
// с автоматическим fallback на in-process очередь при недоступности брокера.
type ExportQueue struct {
	redisClient *redis.Client
	svc         *service.PriceListService
	config      *config.Config

	running    atomic.Bool
	statusMu   sync.RWMutex
	lastStatus domain.ExportStatus
	localQueue chan domain.ExportTask
}

// NewExportQueue создает экземпляр очереди экспорта
func NewExportQueue(redisClient *redis.Client, svc *service.PriceListService, cfg *config.Config) *ExportQueue {
	return &ExportQueue{
		redisClient: redisClient,
		svc:         svc,
		config:      cfg,
		localQueue:  make(chan domain.ExportTask, 10),
		lastStatus: domain.ExportStatus{
			Status: "idle",
		},
	}
}

// IsBusy сообщает, выполняется ли экспорт прямо сейчас
func (q *ExportQueue) IsBusy() bool {
	if q.running.Load() {
		return true
	}

	if q.redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		status := q.GetStatus(ctx)
		if (status.Status == "running" || status.Status == "queued") && status.StartedAt != nil {
			if time.Since(*status.StartedAt) < 30*time.Minute {
				return true
			}
		}
	}

	return false
}

// ResetRunning сбрасывает флаг выполнения (используется при отмене или в тестах)
func (q *ExportQueue) ResetRunning() {
	q.running.Store(false)
}

// GetStatus возвращает статус текущей или последней выполненной задачи
func (q *ExportQueue) GetStatus(ctx context.Context) domain.ExportStatus {
	if q.redisClient != nil {
		val, err := q.redisClient.Get(ctx, exportStatusKey).Result()
		if err == nil {
			var status domain.ExportStatus
			if jsonErr := json.Unmarshal([]byte(val), &status); jsonErr == nil {
				return status
			}
		}
	}

	q.statusMu.RLock()
	defer q.statusMu.RUnlock()
	return q.lastStatus
}

func (q *ExportQueue) setStatus(ctx context.Context, status domain.ExportStatus) {
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
func (q *ExportQueue) Enqueue(ctx context.Context, target string, triggeredBy string) (domain.ExportTask, error) {
	if q.IsBusy() {
		return domain.ExportTask{}, errors.New("экспорт уже выполняется в данный момент")
	}

	if !q.running.CompareAndSwap(false, true) {
		return domain.ExportTask{}, errors.New("экспорт уже выполняется в данный момент")
	}

	task := domain.ExportTask{
		TaskID:      fmt.Sprintf("export-%d", time.Now().UnixNano()),
		Target:      target,
		TriggeredBy: triggeredBy,
		CreatedAt:   time.Now(),
	}

	q.setStatus(ctx, domain.ExportStatus{
		TaskID:      task.TaskID,
		Status:      "queued",
		TriggeredBy: task.TriggeredBy,
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
				slog.InfoContext(ctx, "Задача экспорта опубликована в Redis Streams", slog.String("task_id", task.TaskID))
				return task, nil
			}
			slog.WarnContext(ctx, "Не удалось отправить задачу в Redis Streams, переключение на локальную очередь", logger.Err(err))
		}
	}

	select {
	case q.localQueue <- task:
		slog.InfoContext(ctx, "Задача экспорта поставлена в локальную очередь", slog.String("task_id", task.TaskID))
		return task, nil
	default:
		q.running.Store(false)
		return domain.ExportTask{}, errors.New("очередь задач переполнена")
	}
}

// StartWorker запускает обработку задач
func (q *ExportQueue) StartWorker(ctx context.Context) {
	go q.localWorker(ctx)

	if q.redisClient != nil {
		worker := redisstream.NewConsumerGroupWorker(q.redisClient, q.handleStreamMessage, redisstream.ConsumerConfig{
			Stream:        exportStreamKey,
			Group:         exportGroupKey,
			BatchSize:     1,
			BlockDuration: 5 * time.Second,
		})
		go func() {
			if err := worker.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
				slog.ErrorContext(ctx, "Redis streams worker failed in export-service", logger.Err(err))
			}
		}()
	}
}

func (q *ExportQueue) localWorker(ctx context.Context) {
	slog.InfoContext(ctx, "Запущен локальный in-process воркер экспорта")
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-q.localQueue:
			q.executeTask(ctx, task)
		}
	}
}

func (q *ExportQueue) handleStreamMessage(ctx context.Context, msg redis.XMessage) error {
	payload, ok := msg.Values["payload"].(string)
	if !ok {
		return nil
	}

	var task domain.ExportTask
	if err := json.Unmarshal([]byte(payload), &task); err != nil {
		return nil
	}

	q.executeTask(ctx, task)
	return nil
}

func (q *ExportQueue) executeTask(ctx context.Context, task domain.ExportTask) {
	q.running.Store(true)
	defer q.running.Store(false)

	startedAt := time.Now()
	q.setStatus(ctx, domain.ExportStatus{
		TaskID:      task.TaskID,
		Status:      "running",
		TriggeredBy: task.TriggeredBy,
		StartedAt:   &startedAt,
	})

	slog.InfoContext(ctx, "Начало выполнения задачи экспорта на Drom", slog.String("task_id", task.TaskID))

	xmlData, count, err := q.svc.Generate(ctx)
	if err != nil {
		finishedAt := time.Now()
		errMsg := fmt.Sprintf("ошибка сборки XML: %v", err)
		slog.ErrorContext(ctx, "Сборка XML завершилась неудачей",
			logger.Err(err),
			slog.String("task_id", task.TaskID),
		)
		q.setStatus(ctx, domain.ExportStatus{
			TaskID:      task.TaskID,
			Status:      "failed",
			TriggeredBy: task.TriggeredBy,
			StartedAt:   &startedAt,
			FinishedAt:  &finishedAt,
			Error:       errMsg,
		})
		return
	}

	if err := service.SendToDromAPI(ctx, q.config, xmlData); err != nil {
		finishedAt := time.Now()
		errMsg := fmt.Sprintf("ошибка отправки в Drom API: %v", err)
		slog.ErrorContext(ctx, "Отправка на Drom завершилась неудачей",
			logger.Err(err),
			slog.String("task_id", task.TaskID),
		)
		q.setStatus(ctx, domain.ExportStatus{
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
	slog.InfoContext(ctx, "Экспорт на Drom успешно завершен",
		slog.String("task_id", task.TaskID),
		slog.Int("parts_count", count),
		slog.Duration("duration", time.Since(startedAt)),
	)

	q.setStatus(ctx, domain.ExportStatus{
		TaskID:      task.TaskID,
		Status:      "success",
		TriggeredBy: task.TriggeredBy,
		StartedAt:   &startedAt,
		FinishedAt:  &finishedAt,
		PartsCount:  count,
	})
}
