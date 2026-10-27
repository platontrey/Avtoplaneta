-- name: ListShelves :many
SELECT id, name, created_at
FROM warehouse_shelves
ORDER BY name ASC;

-- name: CreateShelf :one
INSERT INTO warehouse_shelves (name)
VALUES ($1)
ON CONFLICT (LOWER(BTRIM(name))) DO UPDATE
SET name = EXCLUDED.name
RETURNING id, name, created_at;

-- name: DeleteShelf :execrows
DELETE FROM warehouse_shelves
WHERE id = $1;
