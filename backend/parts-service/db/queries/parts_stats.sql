-- name: GetStatsTotals :one
SELECT
    COUNT(*)::bigint AS total_parts,
    COALESCE(SUM(quantity), 0)::bigint AS total_quantity,
    COALESCE(SUM(price * quantity), 0)::float8 AS total_value
FROM parts
WHERE to_delete_at IS NULL AND quantity >= 1 AND deleted_at IS NULL;

-- name: GetStatsTotalsBefore :one
SELECT
    COUNT(*)::bigint AS prev_parts,
    COALESCE(SUM(quantity), 0)::bigint AS prev_quantity,
    COALESCE(SUM(price * quantity), 0)::float8 AS prev_value
FROM parts
WHERE to_delete_at IS NULL
  AND quantity >= 1
  AND created_at < $1
  AND (deleted_at IS NULL OR deleted_at >= $1);

-- name: GetStatsCategories :many
SELECT category AS name, COUNT(*)::bigint AS count
FROM parts
WHERE to_delete_at IS NULL AND quantity >= 1 AND deleted_at IS NULL AND category != ''
GROUP BY category
ORDER BY count DESC;
