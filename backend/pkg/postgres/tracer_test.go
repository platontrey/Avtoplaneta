package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestSQLOperationName(t *testing.T) {
	tests := []struct {
		sql      string
		expected string
	}{
		{"", "db.query"},
		{"   ", "db.query"},
		{"SELECT * FROM users", "db.select"},
		{"  INSERT INTO orders VALUES (1)", "db.insert"},
		{"UPDATE parts SET price = 10", "db.update"},
		{"DELETE FROM sessions", "db.delete"},
	}

	for _, tt := range tests {
		require.Equal(t, tt.expected, sqlOperationName(tt.sql))
	}
}

func TestPostgresTracer_QueryTrace(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)

	tracer := NewTracer()

	ctx := context.Background()
	ctx = tracer.TraceQueryStart(ctx, nil, pgx.TraceQueryStartData{
		SQL: "SELECT id, name FROM users WHERE id = $1",
	})

	tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{
		CommandTag: pgconn.NewCommandTag("SELECT 1"),
		Err:        nil,
	})

	spans := sr.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, "db.select", spans[0].Name())

	var foundSQL bool
	for _, attr := range spans[0].Attributes() {
		if attr.Key == "db.statement" && attr.Value.AsString() == "SELECT id, name FROM users WHERE id = $1" {
			foundSQL = true
		}
	}
	require.True(t, foundSQL)
}

func TestPostgresTracer_QueryError(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)

	tracer := NewTracer()

	ctx := context.Background()
	ctx = tracer.TraceQueryStart(ctx, nil, pgx.TraceQueryStartData{
		SQL: "SELECT * FROM non_existent",
	})

	tracer.TraceQueryEnd(ctx, nil, pgx.TraceQueryEndData{
		Err: errors.New("table does not exist"),
	})

	spans := sr.Ended()
	require.Len(t, spans, 1)
	require.NotEmpty(t, spans[0].Events())
}
