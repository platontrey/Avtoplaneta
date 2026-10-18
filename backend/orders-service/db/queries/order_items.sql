-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, part_id, quantity, price, part_name_snapshot)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetOrderItemsByOrderID :many
SELECT * FROM order_items
WHERE order_id = $1
ORDER BY id ASC;

-- name: GetOrderItemsByOrderIDs :many
SELECT * FROM order_items
WHERE order_id = ANY($1::bigint[])
ORDER BY id ASC;

-- name: FindOrderItem :one
SELECT * FROM order_items
WHERE order_id = $1 AND part_id = $2;

-- name: UpdateOrderItem :exec
UPDATE order_items
SET quantity = $2,
    price = $3
WHERE id = $1;

-- name: DeleteOrderItem :exec
DELETE FROM order_items
WHERE id = $1 AND order_id = $2;

-- name: DeleteOrderItemsByOrderID :exec
DELETE FROM order_items
WHERE order_id = $1;
