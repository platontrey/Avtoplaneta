-- name: GetMessagesByConversationID :many
SELECT * FROM messages
WHERE conversation_id = $1
ORDER BY created_at ASC;

-- name: GetMessageByID :one
SELECT * FROM messages
WHERE id = $1;

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
