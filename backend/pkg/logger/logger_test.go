package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestLoggerJSONFormattingAndServiceAttr(t *testing.T) {
	var buf bytes.Buffer
	l := New(Config{
		Level:       "INFO",
		Format:      FormatJSON,
		ServiceName: "test-service",
		Output:      &buf,
	})

	l.Info("пользователь авторизован", "user_id", 42)

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, "пользователь авторизован", entry["msg"])
	assert.Equal(t, "test-service", entry["service"])
	assert.Equal(t, float64(42), entry["user_id"])
}

func TestLoggerTextFormatting(t *testing.T) {
	var buf bytes.Buffer
	l := New(Config{
		Level:       "WARN",
		Format:      FormatText,
		ServiceName: "worker",
		Output:      &buf,
	})

	l.Warn("предупреждение системы", "attempt", 3)
	output := buf.String()

	assert.Contains(t, output, "level=WARN")
	assert.Contains(t, output, "msg=\"предупреждение системы\"")
	assert.Contains(t, output, "service=worker")
	assert.Contains(t, output, "attempt=3")
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := New(Config{
		Level:  "ERROR",
		Format: FormatJSON,
		Output: &buf,
	})

	l.Debug("отладочное сообщение")
	l.Info("информационное сообщение")
	l.Warn("предупреждение")
	assert.Empty(t, buf.String(), "сообщения ниже уровня ERROR не должны попадать в лог")

	l.Error("критическая ошибка")
	assert.Contains(t, buf.String(), "критическая ошибка")
}

func TestLoggerErrHelper(t *testing.T) {
	testErr := errors.New("соединение разорвано")
	attr := Err(testErr)
	assert.Equal(t, "error", attr.Key)
	assert.Equal(t, testErr, attr.Value.Any())

	nilAttr := Err(nil)
	assert.Equal(t, "error", nilAttr.Key)
	assert.Equal(t, "<nil>", nilAttr.Value.Any())
}

func TestLoggerOpenTelemetryTraceInjection(t *testing.T) {
	var buf bytes.Buffer
	l := New(Config{
		Level:       "INFO",
		Format:      FormatJSON,
		ServiceName: "traced-service",
		Output:      &buf,
	})

	// Создаем моковый контекст с валидным trace.SpanContext
	traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
		Remote:  false,
	})

	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	l.InfoContext(ctx, "запрос с трейсингом", "key", "val")

	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	require.NoError(t, err)

	assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", entry["trace_id"])
	assert.Equal(t, "00f067aa0ba902b7", entry["span_id"])
	assert.Equal(t, "traced-service", entry["service"])
	assert.Equal(t, "val", entry["key"])
}

func TestLoggerInitSetsDefault(t *testing.T) {
	var buf bytes.Buffer
	l := Init(Config{
		Level:       "DEBUG",
		Format:      FormatJSON,
		ServiceName: "default-svc",
		Output:      &buf,
	})
	assert.Equal(t, l, slog.Default())

	slog.Info("тест через slog.Info")
	assert.True(t, strings.Contains(buf.String(), "default-svc"))
}
