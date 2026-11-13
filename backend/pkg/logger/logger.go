package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

// Format формат вывода логов
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Config конфигурация логирования
type Config struct {
	Level       string    // DEBUG, INFO, WARN, ERROR (по умолчанию: INFO)
	Format      Format    // json или text (по умолчанию: json)
	ServiceName string    // Имя сервиса для автоматического добавления атрибута "service"
	Output      io.Writer // Поток вывода (по умолчанию: os.Stdout)
	AddSource   bool      // Включать ли информацию о файле и строке вызова
}

// otelHandler обогащает каждую запись лога trace_id и span_id из OpenTelemetry контекста
type otelHandler struct {
	slog.Handler
}

// Handle добавляет trace_id и span_id к записи лога, если в ctx есть активный спан
func (h *otelHandler) Handle(ctx context.Context, record slog.Record) error {
	if ctx != nil {
		span := trace.SpanFromContext(ctx)
		sc := span.SpanContext()
		if sc.IsValid() {
			record.AddAttrs(
				slog.String("trace_id", sc.TraceID().String()),
				slog.String("span_id", sc.SpanID().String()),
			)
		}
	}
	return h.Handler.Handle(ctx, record)
}

// WithAttrs возвращает новый handler с прикрепленными атрибутами
func (h *otelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &otelHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup возвращает новый handler с именованной группой
func (h *otelHandler) WithGroup(name string) slog.Handler {
	return &otelHandler{Handler: h.Handler.WithGroup(name)}
}

// ParseLevel парсит текстовый уровень логирования в slog.Level
func ParseLevel(lvl string) slog.Level {
	switch strings.ToUpper(strings.TrimSpace(lvl)) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// New создает настроенный *slog.Logger на основе Config с поддержкой OpenTelemetry
func New(cfg Config) *slog.Logger {
	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level:     ParseLevel(cfg.Level),
		AddSource: cfg.AddSource,
	}

	var baseHandler slog.Handler
	if cfg.Format == FormatText {
		baseHandler = slog.NewTextHandler(out, opts)
	} else {
		baseHandler = slog.NewJSONHandler(out, opts)
	}

	handler := slog.Handler(&otelHandler{Handler: baseHandler})

	if cfg.ServiceName != "" {
		handler = handler.WithAttrs([]slog.Attr{slog.String("service", cfg.ServiceName)})
	}

	return slog.New(handler)
}

// Init инициализирует глобальный логгер slog.SetDefault и возвращает его
func Init(cfg Config) *slog.Logger {
	l := New(cfg)
	slog.SetDefault(l)
	return l
}

// Err хелпер для логирования ошибки в виде атрибута slog.Any("error", err)
// аналогично logrus.WithError(err)
func Err(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "<nil>")
	}
	return slog.Any("error", err)
}
