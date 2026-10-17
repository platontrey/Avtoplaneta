-- name: CreateUser :one
INSERT INTO users (username, email, name, initials, inn, provider, role, password)
VALUES (
    sqlc.arg(username)::text,
    sqlc.arg(email)::text,
    sqlc.arg(name)::text,
    sqlc.arg(initials)::text,
    sqlc.arg(inn)::text,
    sqlc.arg(provider)::text,
    sqlc.arg(role)::text,
    sqlc.arg(password)::text
)
RETURNING id, username, email, name, initials, inn, provider, role, password;

-- name: GetUserByID :one
SELECT id, username, email, name, initials, inn, provider, role, password
FROM users
WHERE id = $1;

-- name: GetUserByLoginOrEmail :one
SELECT id, username, email, name, initials, inn, provider, role, password
FROM users
WHERE (username != '' AND LOWER(TRIM(username)) = LOWER(TRIM(sqlc.arg(login)::text)))
   OR (email != '' AND LOWER(TRIM(email)) = LOWER(TRIM(sqlc.arg(login)::text)))
   OR LOWER(TRIM(name)) = LOWER(TRIM(sqlc.arg(login)::text))
   OR (initials != '' AND LOWER(TRIM(initials)) = LOWER(TRIM(sqlc.arg(login)::text)))
LIMIT 1;

-- name: GetUserByEmailOrName :one
SELECT id, username, email, name, initials, inn, provider, role, password
FROM users
WHERE (username != '' AND LOWER(TRIM(username)) = LOWER(TRIM(sqlc.arg(email)::text)))
   OR (email != '' AND LOWER(TRIM(email)) = LOWER(TRIM(sqlc.arg(email)::text)))
   OR LOWER(TRIM(name)) = LOWER(TRIM(sqlc.arg(name)::text))
   OR (initials != '' AND LOWER(TRIM(initials)) = LOWER(TRIM(sqlc.arg(email)::text)))
LIMIT 1;

-- name: UpdateUser :exec
UPDATE users
SET
    username = COALESCE(NULLIF(sqlc.arg(username)::text, ''), username),
    name = COALESCE(NULLIF(sqlc.arg(name)::text, ''), name),
    email = COALESCE(NULLIF(sqlc.arg(email)::text, ''), email),
    initials = COALESCE(NULLIF(sqlc.arg(initials)::text, ''), initials),
    inn = COALESCE(NULLIF(sqlc.arg(inn)::text, ''), inn),
    role = COALESCE(NULLIF(sqlc.arg(role)::text, ''), role)
WHERE id = sqlc.arg(id);

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: FindAllUsers :many
SELECT id, username, email, name, initials, inn, provider, role, password
FROM users
ORDER BY id;

-- name: ExistsByUsername :one
SELECT COUNT(*) > 0 AS exists FROM users 
WHERE username != '' AND LOWER(TRIM(username)) = LOWER(TRIM(sqlc.arg(username)::text));

-- name: ExistsByEmail :one
SELECT COUNT(*) > 0 AS exists FROM users 
WHERE email != '' AND LOWER(TRIM(email)) = LOWER(TRIM(sqlc.arg(email)::text));

-- name: ExistsByName :one
SELECT COUNT(*) > 0 AS exists FROM users 
WHERE LOWER(TRIM(name)) = LOWER(TRIM(sqlc.arg(name)::text));

-- name: CountAllUsers :one
SELECT COUNT(*)::BIGINT AS count FROM users;
