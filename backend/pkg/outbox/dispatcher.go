package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// Event представляет единицу события Transactional Outbox
type Event struct {
	ID            int64  `json:"id"`
	AggregateType string `json:"aggregate_type"`
	AggregateID   int64  `json:"aggregate_id"`
	EventType     string `json:"event_type"`
	Payload       []byte `json:"payload"`
	RetryCount    int32  `json:"retry_count"`
}

// Storage интерфейс работы с таблицей outbox_events
type Storage interface {
	FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]Event, error)
	MarkOutboxEventProcessed(ctx context.Context, id int64) error
	MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error
	CleanupProcessedOutboxEvents(ctx context.Context) error
}

// Handler функция обработки отдельного outbox-события
type Handler func(ctx context.Context, event Event) error

// MetricsRecorder интерфейс для записи Prometheus метрик диспетчера
type MetricsRecorder interface {
	RecordEventProcessed(eventType string, durationSec float64)
	RecordEventFailed(eventType string, durationSec float64)
}

// PrometheusMetrics адаптер для стандартных Prometheus векторов сервиса
type PrometheusMetrics struct {
	Processed interface {
		WithLabelValues(lvs ...string) prometheus.Counter
	}
	Failed interface {
		WithLabelValues(lvs ...string) prometheus.Counter
	}
	Duration interface {
		WithLabelValues(lvs ...string) prometheus.Observer
	}
}

// RecordEventProcessed записывает успешную обработку события
func (m PrometheusMetrics) RecordEventProcessed(eventType string, durationSec float64) {
	if m.Processed != nil {
		m.Processed.WithLabelValues(eventType).Inc()
	}
	if m.Duration != nil {
		m.Duration.WithLabelValues(eventType).Observe(durationSec)
	}
}

// RecordEventFailed записывает ошибку обработки события
func (m PrometheusMetrics) RecordEventFailed(eventType string, durationSec float64) {
	if m.Failed != nil {
		m.Failed.WithLabelValues(eventType).Inc()
	}
	if m.Duration != nil {
		m.Duration.WithLabelValues(eventType).Observe(durationSec)
	}
}

// Options конфигурационные параметры диспетчера Outbox
type Options struct {
	Name                 string          // Название сервиса для логирования (например, "Parts", "Orders", "Auth")
	Channel              string          // Канал PostgreSQL LISTEN (например, "outbox_events_channel")
	BatchSize            int32           // Размер пачки выборки (default: 50)
	PollInterval         time.Duration   // Интервал поллинга при отсутствии LISTEN (default: 100ms)
	FallbackPollInterval time.Duration   // Страховочный интервал при активном LISTEN (default: 3s)
	CleanupInterval      time.Duration   // Период запуска очистки старых событий (default: 1h)
	Metrics              MetricsRecorder // Опциональный сборщик метрик
	Logger               *slog.Logger    // Опциональный логгер с контекстом сервиса
}

// Dispatcher управляет надежным циклом Transactional Outbox (LISTEN/NOTIFY, поллинг, обработка, retry, cleanup)
type Dispatcher struct {
	pool    *pgxpool.Pool
	storage Storage
	handler Handler
	opts    Options
	log     *slog.Logger
}

// NewDispatcher создает новый экземпляр диспетчера Outbox
func NewDispatcher(pool *pgxpool.Pool, storage Storage, handler Handler, opts Options) *Dispatcher {
	if opts.BatchSize <= 0 {
		opts.BatchSize = 50
	}
	if opts.PollInterval <= 0 {
		opts.PollInterval = 100 * time.Millisecond
	}
	if opts.FallbackPollInterval <= 0 {
		opts.FallbackPollInterval = 3 * time.Second
	}
	if opts.CleanupInterval <= 0 {
		opts.CleanupInterval = 1 * time.Hour
	}
	if opts.Name == "" {
		opts.Name = "Outbox"
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.Default().With(slog.String("component", opts.Name+"-outbox"))
	}

	return &Dispatcher{
		pool:    pool,
		storage: storage,
		handler: handler,
		opts:    opts,
		log:     logger,
	}
}

// Start запускает фоновый воркер очистки, слушатель LISTEN/NOTIFY и основной цикл обработки
func (d *Dispatcher) Start(ctx context.Context) error {
	d.log.InfoContext(ctx, "Starting Transactional Outbox dispatcher", slog.String("name", d.opts.Name))

	go d.startCleanupLoop(ctx)

	notifyCh := make(chan struct{}, 1)
	if d.pool != nil && d.opts.Channel != "" {
		go d.startNotifyListener(ctx, notifyCh)
	}

	interval := d.opts.PollInterval
	if d.pool != nil && d.opts.Channel != "" {
		interval = d.opts.FallbackPollInterval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Первоначальный прогон при старте
	d.ProcessPendingEvents(ctx)

	for {
		select {
		case <-ctx.Done():
			d.log.InfoContext(ctx, "Stopping Transactional Outbox dispatcher", slog.String("name", d.opts.Name))
			return nil

		case <-notifyCh:
			d.ProcessPendingEvents(ctx)

		case <-ticker.C:
			d.ProcessPendingEvents(ctx)
		}
	}
}

// startNotifyListener слушает PostgreSQL LISTEN на указанном канале с автоматическим переподключением
func (d *Dispatcher) startNotifyListener(ctx context.Context, notifyCh chan<- struct{}) {
	for {
		if ctx.Err() != nil {
			return
		}

		conn, err := d.pool.Acquire(ctx)
		if err != nil {
			d.log.WarnContext(ctx, "Failed to acquire connection for LISTEN, retrying in 2s", slog.Any("error", err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		_, err = conn.Exec(ctx, "LISTEN "+d.opts.Channel)
		if err != nil {
			conn.Release()
			d.log.WarnContext(ctx, "Failed to execute LISTEN, retrying in 2s", slog.Any("error", err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		d.log.InfoContext(ctx, "Successfully subscribed via PostgreSQL LISTEN", slog.String("channel", d.opts.Channel))

		for {
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				conn.Release()
				if ctx.Err() != nil {
					return
				}
				d.log.WarnContext(ctx, "Connection lost while waiting for notification, reconnecting...", slog.Any("error", err))
				break
			}

			if notification != nil {
				select {
				case notifyCh <- struct{}{}:
				default:
					// Сигнал уже в очереди
				}
			}
		}
	}
}

// startCleanupLoop периодически удаляет старые обработанные события
func (d *Dispatcher) startCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(d.opts.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.storage.CleanupProcessedOutboxEvents(ctx); err != nil {
				d.log.WarnContext(ctx, "Failed to cleanup processed outbox events", slog.Any("error", err))
			} else {
				d.log.DebugContext(ctx, "Cleaned up old processed outbox events")
			}
		}
	}
}

// ProcessPendingEvents вычитывает и обрабатывает пачки событий до опустошения очереди
func (d *Dispatcher) ProcessPendingEvents(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		events, err := d.storage.FetchPendingOutboxEvents(ctx, d.opts.BatchSize)
		if err != nil {
			d.log.ErrorContext(ctx, "Failed to fetch pending events", slog.Any("error", err))
			return
		}

		if len(events) == 0 {
			return
		}

		for _, event := range events {
			if ctx.Err() != nil {
				return
			}

			start := time.Now()
			err := d.handler(ctx, event)
			duration := time.Since(start).Seconds()

			if err != nil {
				if d.opts.Metrics != nil {
					d.opts.Metrics.RecordEventFailed(event.EventType, duration)
				}
				d.log.ErrorContext(ctx, "Failed to process outbox event",
					slog.Any("error", err),
					slog.Int64("event_id", event.ID),
					slog.String("event_type", event.EventType),
					slog.Int64("aggregate_id", event.AggregateID),
					slog.String("aggregate_type", event.AggregateType),
				)

				if markErr := d.storage.MarkOutboxEventFailed(ctx, event.ID, err.Error()); markErr != nil {
					d.log.ErrorContext(ctx, "Failed to mark event as failed",
						slog.Any("error", markErr),
						slog.Int64("event_id", event.ID),
					)
				}
			} else {
				if d.opts.Metrics != nil {
					d.opts.Metrics.RecordEventProcessed(event.EventType, duration)
				}
				if markErr := d.storage.MarkOutboxEventProcessed(ctx, event.ID); markErr != nil {
					d.log.ErrorContext(ctx, "Failed to mark event as processed",
						slog.Any("error", markErr),
						slog.Int64("event_id", event.ID),
					)
				}
			}
		}

		// Если пачка меньше лимита, значит очередь исчерпана
		if len(events) < int(d.opts.BatchSize) {
			return
		}
	}
}
