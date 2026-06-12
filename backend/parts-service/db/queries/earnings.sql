-- name: GetEarnings :one
SELECT * FROM earnings ORDER BY id LIMIT 1;

-- name: UpdateEarnings :exec
UPDATE earnings SET total_amount = $1, updated_at = NOW() WHERE id = $2;

-- name: UpsertEarnings :one
INSERT INTO earnings (total_amount, updated_at)
VALUES ($1, NOW())
ON CONFLICT (id) DO UPDATE SET total_amount = $1, updated_at = NOW()
RETURNING *;

-- name: CreateEarnings :one
INSERT INTO earnings (total_amount, updated_at) VALUES ($1, NOW()) RETURNING *;
