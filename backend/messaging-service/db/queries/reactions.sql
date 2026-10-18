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
