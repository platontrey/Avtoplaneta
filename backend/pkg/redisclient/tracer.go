package redisclient

import (
	"context"
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type otelRedisHook struct {
	tracer trace.Tracer
}

// NewHook создает новый OpenTelemetry хук для go-redis
func NewHook() redis.Hook {
	return &otelRedisHook{
		tracer: otel.Tracer("avtoplaneta/pkg/redisclient"),
	}
}

// Instrument подключает трассировку OpenTelemetry к переданному redis.Client
func Instrument(client *redis.Client) *redis.Client {
	if client != nil {
		client.AddHook(NewHook())
	}
	return client
}

func (h *otelRedisHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *otelRedisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		cmdName := cmd.Name()
		if cmdName == "" {
			cmdName = "redis"
		}
		spanName := "redis." + strings.ToLower(cmdName)

		ctx, span := h.tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("db.system", "redis"),
				attribute.String("db.operation", cmdName),
				attribute.String("db.statement", cmd.String()),
			),
		)
		defer span.End()

		err := next(ctx, cmd)
		if err != nil && !errors.Is(err, redis.Nil) {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		return err
	}
}

func (h *otelRedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		spanName := "redis.pipeline"
		ctx, span := h.tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("db.system", "redis"),
				attribute.Int("db.redis.num_cmd", len(cmds)),
			),
		)
		defer span.End()

		err := next(ctx, cmds)
		if err != nil && !errors.Is(err, redis.Nil) {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}
		return err
	}
}
