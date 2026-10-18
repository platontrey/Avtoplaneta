-- name: GetUserStatuses :many
SELECT * FROM user_status;

-- name: GetUserStatusesByUserIDs :many
SELECT * FROM user_status
WHERE user_id = ANY($1::bigint[]);

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
