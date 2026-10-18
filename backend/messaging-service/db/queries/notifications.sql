-- name: GetNotificationsByUserID :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateNotification :one
INSERT INTO notifications (user_id, type, message, read, created_at, related_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: MarkNotificationRead :exec
UPDATE notifications
SET read = TRUE
WHERE id = $1;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read = TRUE
WHERE user_id = $1;
