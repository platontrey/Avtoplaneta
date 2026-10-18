-- name: DecreasePartQuantity :exec
UPDATE parts
SET quantity = CASE WHEN quantity - sqlc.arg(amount)::integer <= 0 THEN -1 ELSE quantity - sqlc.arg(amount)::integer END
WHERE id = sqlc.arg(id);

-- name: IncreasePartQuantity :exec
UPDATE parts
SET quantity = CASE WHEN quantity = -1 THEN sqlc.arg(amount)::integer ELSE quantity + sqlc.arg(amount)::integer END
WHERE id = sqlc.arg(id);

-- name: RecordStockOperation :execrows
INSERT INTO part_stock_operations (operation_id, part_id, operation_type, amount)
VALUES ($1, $2, $3, $4)
ON CONFLICT (operation_id, part_id, operation_type) DO NOTHING;

-- name: GetInventoryVersion :one
SELECT MAX(updated_at)::timestamptz AS last_change, COUNT(*)::bigint AS alive
FROM parts
WHERE deleted_at IS NULL;
