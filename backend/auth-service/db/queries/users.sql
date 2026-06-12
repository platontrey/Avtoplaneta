-- name: CreateUser :one
INSERT INTO users (email, name, initials, inn, provider, role, password)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, email, name, initials, inn, provider, role, password;

-- name: GetUserByID :one
SELECT id, email, name, initials, inn, provider, role, password
FROM users
WHERE id = $1;

-- name: GetUserByEmailOrName :one
SELECT id, email, name, initials, inn, provider, role, password
FROM users
WHERE email = $1 OR name = $2
LIMIT 1;

-- name: UpdateUser :exec
UPDATE users
SET
    name = COALESCE(NULLIF($2::text, ''), name),
    email = COALESCE(NULLIF($3::text, ''), email),
    initials = COALESCE(NULLIF($4::text, ''), initials),
    inn = COALESCE(NULLIF($5::text, ''), inn),
    role = COALESCE(NULLIF($6::text, ''), role)
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: FindAllUsers :many
SELECT id, email, name, initials, inn, provider, role, password
FROM users
ORDER BY id;

-- name: ExistsByEmail :one
SELECT COUNT(*) > 0 AS exists FROM users WHERE email = $1;

-- name: ExistsByName :one
SELECT COUNT(*) > 0 AS exists FROM users WHERE name = $1;

-- name: CountAllUsers :one
SELECT COUNT(*)::BIGINT AS count FROM users;
