package repository

import (
	"github.com/jackc/pgx/v5/pgtype"
	"context"

	"parts-service/db/sqlc"
)

// ─── Outbox ─────────────────────────────────────────────────────────────────

func (r *partRepository) FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]sqlc.OutboxEvent, error) {
	return r.queries.FetchPendingOutboxEvents(ctx, limit)
}

func (r *partRepository) MarkOutboxEventProcessed(ctx context.Context, id int64) error {
	return r.queries.MarkOutboxEventProcessed(ctx, id)
}

func (r *partRepository) MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error {
	return r.queries.MarkOutboxEventFailed(ctx, sqlc.MarkOutboxEventFailedParams{
		ID:        id,
		LastError: pgtype.Text{String: lastErr, Valid: lastErr != ""},
	})
}

func (r *partRepository) CleanupProcessedOutboxEvents(ctx context.Context) error {
	return r.queries.CleanupProcessedOutboxEvents(ctx)
}
