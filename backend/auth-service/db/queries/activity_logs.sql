-- name: CreateActivityLog :one
INSERT INTO user_activity_logs (user_id, user_name, user_email, action, resource_type, resource_id, details, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, user_name, user_email, action, resource_type, resource_id, details, ip_address, user_agent, created_at;
