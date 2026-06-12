-- name: GetConversations :many
SELECT c.*, COALESCE(u.unread_count, 0)::bigint as unread_count
FROM conversations c
LEFT JOIN (
    SELECT m.conversation_id, COUNT(*) as unread_count
    FROM messages m
    WHERE NOT (m.read_by @> ARRAY[sqlc.arg(user_id)::bigint])
      AND m.sender_id != sqlc.arg(user_id)::bigint
    GROUP BY m.conversation_id
) u ON u.conversation_id = c.id
WHERE c.participants @> ARRAY[sqlc.arg(user_id)::bigint]
ORDER BY c.last_message_at DESC;

-- name: GetConversation :one
SELECT * FROM conversations
WHERE id = $1;

-- name: CreateConversation :one
INSERT INTO conversations (participants, title, created_at, last_message_at, last_message)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateConversation :one
UPDATE conversations
SET title = $2,
    last_message_at = $3,
    last_message = $4,
    participants = $5
WHERE id = $1
RETURNING *;

-- name: UpdateConversationLastMessage :exec
UPDATE conversations
SET last_message_at = $2,
    last_message = $3
WHERE id = $1;

-- name: DeleteConversation :exec
DELETE FROM conversations
WHERE id = $1;

-- name: GetMessagesByConversationID :many
SELECT * FROM messages
WHERE conversation_id = $1
ORDER BY created_at ASC;

-- name: CreateMessage :one
INSERT INTO messages (conversation_id, sender_id, content, message_type, voice_url, created_at, read_by, attachments)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: DeleteMessage :exec
DELETE FROM messages
WHERE id = $1;

-- name: DeleteMessagesByConversationID :exec
DELETE FROM messages
WHERE conversation_id = $1;

-- name: AddUserToReadBy :exec
UPDATE messages
SET read_by = array_append(read_by, sqlc.arg(user_id)::bigint)
WHERE id = sqlc.arg(message_id) AND NOT (read_by @> ARRAY[sqlc.arg(user_id)::bigint]);

-- name: GetReactionByID :one
SELECT * FROM reactions
WHERE id = $1;

-- name: CreateReaction :one
INSERT INTO reactions (message_id, user_id, emoji)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteReaction :exec
DELETE FROM reactions
WHERE id = $1;

-- name: DeleteReactionByFields :exec
DELETE FROM reactions
WHERE message_id = $1 AND user_id = $2 AND emoji = $3;

-- name: GetReactionsByMessageIDs :many
SELECT * FROM reactions
WHERE message_id = ANY($1::bigint[]);

-- name: GetMentionsByMessageIDs :many
SELECT * FROM mentions
WHERE message_id = ANY($1::bigint[]);

-- name: CreateMention :one
INSERT INTO mentions (message_id, type, resource_id, text)
VALUES ($1, $2, $3, $4)
RETURNING *;

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

-- name: GetUserStatuses :many
SELECT * FROM user_status;

-- name: GetUserStatusByUserID :one
SELECT * FROM user_status
WHERE user_id = $1;

-- name: UpsertUserStatus :one
INSERT INTO user_status (user_id, is_online, last_seen)
VALUES ($1, $2, $3)
ON CONFLICT (user_id)
DO UPDATE SET is_online = EXCLUDED.is_online,
              last_seen = EXCLUDED.last_seen
RETURNING *;

-- name: SearchMessagesWithUser :many
SELECT m.* FROM messages m
JOIN conversations c ON m.conversation_id = c.id
WHERE c.participants @> ARRAY[sqlc.arg(user_id)::bigint]
  AND (m.search_vector @@ plainto_tsquery('russian', sqlc.arg(query)::text) OR m.content ILIKE '%' || sqlc.arg(query)::text || '%')
ORDER BY m.created_at DESC
LIMIT $1 OFFSET $2;

-- name: SearchMessagesAll :many
SELECT m.* FROM messages m
WHERE (m.search_vector @@ plainto_tsquery('russian', sqlc.arg(query)::text) OR m.content ILIKE '%' || sqlc.arg(query)::text || '%')
ORDER BY m.created_at DESC
LIMIT $1 OFFSET $2;
