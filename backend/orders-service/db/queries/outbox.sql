-- name: CreateOrdersOutboxEvent :one
INSERT INTO outbox_events (
    aggregate_type,
    aggregate_id,
    event_type,
    payload
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: FetchPendingOrdersOutboxEvents :many
SELECT * FROM outbox_events
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: MarkOrdersOutboxEventProcessed :exec
UPDATE outbox_events
SET status = 'processed',
    processed_at = NOW(),
    last_error = NULL
WHERE id = $1;

-- name: MarkOrdersOutboxEventFailed :exec
UPDATE outbox_events
SET status = 'failed',
    retry_count = retry_count + 1,
    last_error = $2
WHERE id = $1;

-- name: CleanupProcessedOrdersOutboxEvents :exec
DELETE FROM outbox_events
WHERE status = 'processed'
  AND processed_at < NOW() - INTERVAL '24 hours';
