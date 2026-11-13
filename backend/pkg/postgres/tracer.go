package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type pgxTracerKey struct{}

var tracerKey = pgxTracerKey{}

type tracerContextState struct {
	span trace.Span
}

// Tracer реализует трассировку запросов pgx через OpenTelemetry
type Tracer struct {
	tracer trace.Tracer
}

// NewTracer создает новый трассировщик OpenTelemetry для pgx
func NewTracer() *Tracer {
	return &Tracer{
		tracer: otel.Tracer("avtoplaneta/pkg/postgres"),
	}
}

func sqlOperationName(sql string) string {
	clean := strings.TrimSpace(sql)
	if len(clean) == 0 {
		return "db.query"
	}
	fields := strings.Fields(clean)
	if len(fields) > 0 {
		op := strings.ToUpper(fields[0])
		return "db." + strings.ToLower(op)
	}
	return "db.query"
}

// TraceQueryStart вызывается в начале выполнения SQL запроса (Query, QueryRow, Exec)
func (t *Tracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	spanName := sqlOperationName(data.SQL)
	attrs := []attribute.KeyValue{
		attribute.String("db.system", "postgresql"),
		attribute.String("db.statement", data.SQL),
	}
	if conn != nil && conn.Config() != nil {
		if conn.Config().Database != "" {
			attrs = append(attrs, attribute.String("db.name", conn.Config().Database))
		}
		if conn.Config().User != "" {
			attrs = append(attrs, attribute.String("db.user", conn.Config().User))
		}
		if conn.Config().Host != "" {
			attrs = append(attrs, attribute.String("net.peer.name", conn.Config().Host))
		}
		if conn.Config().Port != 0 {
			attrs = append(attrs, attribute.Int64("net.peer.port", int64(conn.Config().Port)))
		}
	}

	ctx, span := t.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
	return context.WithValue(ctx, tracerKey, &tracerContextState{span: span})
}

// TraceQueryEnd вызывается по завершению выполнения SQL запроса
func (t *Tracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	state, ok := ctx.Value(tracerKey).(*tracerContextState)
	if !ok || state == nil || state.span == nil {
		return
	}
	defer state.span.End()

	if data.Err != nil && !errors.Is(data.Err, pgx.ErrNoRows) {
		state.span.RecordError(data.Err)
		state.span.SetStatus(codes.Error, data.Err.Error())
	} else {
		state.span.SetStatus(codes.Ok, "")
	}

	if rows := data.CommandTag.RowsAffected(); rows > 0 {
		state.span.SetAttributes(attribute.Int64("db.rows_affected", rows))
	}
}

// TraceBatchStart начинает спан пакета запросов SendBatch
func (t *Tracer) TraceBatchStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchStartData) context.Context {
	ctx, span := t.tracer.Start(ctx, "db.batch",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
		),
	)
	return context.WithValue(ctx, tracerKey, &tracerContextState{span: span})
}

// TraceBatchQuery вызывается для каждого подзапроса внутри пакета
func (t *Tracer) TraceBatchQuery(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchQueryData) {
	// Подзапросы регистрируются внутри основного спана пакета
}

// TraceBatchEnd завершает спан пакета запросов SendBatch
func (t *Tracer) TraceBatchEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchEndData) {
	state, ok := ctx.Value(tracerKey).(*tracerContextState)
	if !ok || state == nil || state.span == nil {
		return
	}
	defer state.span.End()

	if data.Err != nil && !errors.Is(data.Err, pgx.ErrNoRows) {
		state.span.RecordError(data.Err)
		state.span.SetStatus(codes.Error, data.Err.Error())
	} else {
		state.span.SetStatus(codes.Ok, "")
	}
}

// TracePrepareStart начинает спан подготовки запроса
func (t *Tracer) TracePrepareStart(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareStartData) context.Context {
	ctx, span := t.tracer.Start(ctx, "db.prepare",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", data.SQL),
			attribute.String("db.statement_name", data.Name),
		),
	)
	return context.WithValue(ctx, tracerKey, &tracerContextState{span: span})
}

// TracePrepareEnd завершает спан подготовки запроса
func (t *Tracer) TracePrepareEnd(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareEndData) {
	state, ok := ctx.Value(tracerKey).(*tracerContextState)
	if !ok || state == nil || state.span == nil {
		return
	}
	defer state.span.End()

	if data.Err != nil {
		state.span.RecordError(data.Err)
		state.span.SetStatus(codes.Error, data.Err.Error())
	} else {
		state.span.SetStatus(codes.Ok, "")
	}
}
