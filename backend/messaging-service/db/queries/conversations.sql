-- name: GetConversations :many
SELECT c.*, COALESCE(u.unread_count, 0)::bigint as unread_count
FROM conversations c
LEFT JOIN LATERAL (
    SELECT COUNT(*) as unread_count
    FROM messages m
    WHERE m.conversation_id = c.id
      AND m.sender_id != sqlc.arg(user_id)::bigint
      AND NOT (m.read_by @> ARRAY[sqlc.arg(user_id)::bigint])
) u ON TRUE
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
