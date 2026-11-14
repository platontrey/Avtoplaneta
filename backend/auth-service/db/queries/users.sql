-- ============================================================================
-- Domain: Auth Service / Users
-- Назначение: CRUD операции учетных записей сотрудников, поиск по логину/email,
--             проверка дубликатов и обновление профиля/паролей/аватаров.
-- Генерация Go-кода: cd backend/auth-service && sqlc generate
-- ============================================================================

-- name: CreateUser :one
-- Создает нового пользователя в системе с указанием роли и провайдера аутентификации.
INSERT INTO users (username, email, name, initials, inn, provider, role, password, avatar_url)
VALUES (
    sqlc.arg(username)::text,
    sqlc.arg(email)::text,
    sqlc.arg(name)::text,
    sqlc.arg(initials)::text,
    sqlc.arg(inn)::text,
    sqlc.arg(provider)::text,
    sqlc.arg(role)::text,
    sqlc.arg(password)::text,
    COALESCE(sqlc.narg(avatar_url)::text, '')
)
RETURNING id, username, email, name, initials, inn, provider, role, password, avatar_url;

-- name: GetUserByID :one
-- Возвращает пользователя по его первичному 64-битному идентификатору.
SELECT id, username, email, name, initials, inn, provider, role, password, avatar_url
FROM users
WHERE id = $1;

-- name: GetUserByLoginOrEmail :one
-- Выполняет поиск пользователя для формы входа по логину, email, ФИО или инициалам
-- с нормализацией регистра и очисткой крайних пробелов (LOWER(TRIM(...))).
SELECT id, username, email, name, initials, inn, provider, role, password, avatar_url
FROM users
WHERE (username != '' AND LOWER(TRIM(username)) = LOWER(TRIM(sqlc.arg(login)::text)))
   OR (email != '' AND LOWER(TRIM(email)) = LOWER(TRIM(sqlc.arg(login)::text)))
   OR LOWER(TRIM(name)) = LOWER(TRIM(sqlc.arg(login)::text))
   OR (initials != '' AND LOWER(TRIM(initials)) = LOWER(TRIM(sqlc.arg(login)::text)))
LIMIT 1;

-- name: GetUserByEmailOrName :one
-- Ищет пользователя по совпадению email, username или имени.
SELECT id, username, email, name, initials, inn, provider, role, password, avatar_url
FROM users
WHERE (username != '' AND LOWER(TRIM(username)) = LOWER(TRIM(sqlc.arg(email)::text)))
   OR (email != '' AND LOWER(TRIM(email)) = LOWER(TRIM(sqlc.arg(email)::text)))
   OR LOWER(TRIM(name)) = LOWER(TRIM(sqlc.arg(name)::text))
   OR (initials != '' AND LOWER(TRIM(initials)) = LOWER(TRIM(sqlc.arg(email)::text)))
LIMIT 1;

-- name: UpdateUser :exec
-- Административное обновление всех основных полей пользователя.
-- Пустые строки ('') игнорируются, сохраняя предыдущее значение.
UPDATE users
SET
    username = COALESCE(NULLIF(sqlc.arg(username)::text, ''), username),
    name = COALESCE(NULLIF(sqlc.arg(name)::text, ''), name),
    email = COALESCE(NULLIF(sqlc.arg(email)::text, ''), email),
    initials = COALESCE(NULLIF(sqlc.arg(initials)::text, ''), initials),
    inn = COALESCE(NULLIF(sqlc.arg(inn)::text, ''), inn),
    role = COALESCE(NULLIF(sqlc.arg(role)::text, ''), role),
    avatar_url = COALESCE(sqlc.narg(avatar_url)::text, avatar_url)
WHERE id = sqlc.arg(id);

-- name: UpdateUserProfile :exec
-- Пользовательское обновление личного профиля (ФИО, username, ИНН).
UPDATE users
SET
    name = COALESCE(NULLIF(sqlc.arg(name)::text, ''), name),
    username = COALESCE(NULLIF(sqlc.arg(username)::text, ''), username),
    inn = COALESCE(NULLIF(sqlc.arg(inn)::text, ''), inn)
WHERE id = sqlc.arg(id);

-- name: UpdateUserPassword :exec
-- Обновляет захэшированный пароль пользователя по его ID.
UPDATE users
SET password = sqlc.arg(password)::text
WHERE id = sqlc.arg(id);

-- name: UpdateUserAvatar :exec
-- Обновляет URL аватара пользователя.
UPDATE users
SET avatar_url = sqlc.arg(avatar_url)::text
WHERE id = sqlc.arg(id);

-- name: DeleteUser :exec
-- Физически удаляет пользователя по ID (для мягкого удаления см. deleted_at).
DELETE FROM users WHERE id = $1;

-- name: FindAllUsers :many
-- Возвращает список всех пользователей системы, отсортированный по ID.
SELECT id, username, email, name, initials, inn, provider, role, password, avatar_url
FROM users
ORDER BY id;

-- name: ExistsByUsername :one
-- Проверяет занятость логина (username) без учета регистра.
SELECT COUNT(*) > 0 AS exists FROM users 
WHERE username != '' AND LOWER(TRIM(username)) = LOWER(TRIM(sqlc.arg(username)::text));

-- name: ExistsByEmail :one
-- Проверяет занятость email без учета регистра.
SELECT COUNT(*) > 0 AS exists FROM users 
WHERE email != '' AND LOWER(TRIM(email)) = LOWER(TRIM(sqlc.arg(email)::text));

-- name: ExistsByName :one
-- Проверяет существование пользователя с точным именем (ФИО) без учета регистра.
SELECT COUNT(*) > 0 AS exists FROM users 
WHERE LOWER(TRIM(name)) = LOWER(TRIM(sqlc.arg(name)::text));

-- name: CountAllUsers :one
-- Возвращает общее количество зарегистрированных пользователей.
SELECT COUNT(*)::BIGINT AS count FROM users;
