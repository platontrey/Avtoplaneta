-- name: CreateOutboxEvent :one
INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload, status)
VALUES ($1, $2, $3, $4, 'pending')
RETURNING *;

-- name: FetchPendingOutboxEvents :many
SELECT * FROM outbox_events
WHERE status = 'pending'
ORDER BY id ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: MarkOutboxEventProcessed :exec
UPDATE outbox_events
SET status = 'processed', processed_at = NOW(), last_error = NULL
WHERE id = $1;

-- name: MarkOutboxEventFailed :exec
UPDATE outbox_events
SET status = CASE WHEN retry_count >= 5 THEN 'failed' ELSE 'pending' END,
    retry_count = retry_count + 1,
    last_error = $2
WHERE id = $1;

-- name: CleanupProcessedOutboxEvents :exec
DELETE FROM outbox_events
WHERE status = 'processed' AND processed_at < NOW() - INTERVAL '7 days';
