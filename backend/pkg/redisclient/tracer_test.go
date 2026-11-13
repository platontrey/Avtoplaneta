package redisclient

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestOtelRedisHook_ProcessHook(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)

	hook := NewHook()

	cmd := redis.NewStringCmd(context.Background(), "get", "test-key")
	calledNext := false

	dummyNext := func(ctx context.Context, cmd redis.Cmder) error {
		calledNext = true
		return nil
	}

	wrapped := hook.ProcessHook(dummyNext)
	err := wrapped(context.Background(), cmd)
	require.NoError(t, err)
	require.True(t, calledNext)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, "redis.get", spans[0].Name())

	var foundOp, foundSys bool
	for _, attr := range spans[0].Attributes() {
		if attr.Key == "db.system" && attr.Value.AsString() == "redis" {
			foundSys = true
		}
		if attr.Key == "db.operation" && attr.Value.AsString() == "get" {
			foundOp = true
		}
	}
	require.True(t, foundSys)
	require.True(t, foundOp)
}

func TestOtelRedisHook_ProcessPipelineHook(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)

	hook := NewHook()

	cmds := []redis.Cmder{
		redis.NewStringCmd(context.Background(), "get", "key1"),
		redis.NewStringCmd(context.Background(), "set", "key2", "val2"),
	}
	calledNext := false

	dummyNext := func(ctx context.Context, cmds []redis.Cmder) error {
		calledNext = true
		return nil
	}

	wrapped := hook.ProcessPipelineHook(dummyNext)
	err := wrapped(context.Background(), cmds)
	require.NoError(t, err)
	require.True(t, calledNext)

	spans := sr.Ended()
	require.Len(t, spans, 1)
	require.Equal(t, "redis.pipeline", spans[0].Name())
}
