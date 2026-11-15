-- name: GetItemByID :one
SELECT id, created_at, updated_at, deleted_at, status, name, description, metadata
FROM items
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: ListItems :many
SELECT id, created_at, updated_at, deleted_at, status, name, description, metadata
FROM items
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count)::integer
OFFSET sqlc.arg(offset_count)::integer;

-- name: CreateItem :one
INSERT INTO items (
    name, description, status, metadata, created_at, updated_at
) VALUES (
    sqlc.arg(name),
    sqlc.arg(description),
    sqlc.arg(status),
    sqlc.arg(metadata)::jsonb,
    NOW(),
    NOW()
)
RETURNING id, created_at, updated_at, deleted_at, status, name, description, metadata;

-- name: UpdateItem :one
UPDATE items
SET 
    name = sqlc.arg(name),
    description = sqlc.arg(description),
    status = sqlc.arg(status),
    metadata = sqlc.arg(metadata)::jsonb,
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING id, created_at, updated_at, deleted_at, status, name, description, metadata;

-- name: SoftDeleteItem :execrows
UPDATE items
SET deleted_at = NOW()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;
