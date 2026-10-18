-- name: CreateOrder :one
INSERT INTO orders (
    customer_id, seller_id, seller, part, part_id, location, buyer_number,
    status, status_text, auto_deleted, created_at,
    order_number, source, payment_status, warehouse_status, delivery_method,
    transport_company, tracking_number, notes, discount, completed_at, updated_at
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11,
    $12, $13, $14, $15, $16,
    $17, $18, $19, $20, $21, $22
)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1;

-- name: FindAllOrders :many
SELECT * FROM orders
WHERE auto_deleted = FALSE
ORDER BY created_at DESC;

-- name: FindCompletedOrders :many
SELECT * FROM orders
WHERE auto_deleted = TRUE
ORDER BY COALESCE(completed_at, created_at) DESC
LIMIT 200;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2,
    status_text = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: CompleteOrderRecord :exec
UPDATE orders
SET status = 'green',
    status_text = 'Завершён',
    payment_status = 'paid',
    warehouse_status = 'ready',
    auto_deleted = TRUE,
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteOrder :exec
DELETE FROM orders
WHERE id = $1;

-- name: MarkExpiredAsAutoDeleted :exec
UPDATE orders
SET auto_deleted = TRUE
WHERE created_at <= $1 AND auto_deleted = FALSE;

-- name: UpdateSellerName :exec
UPDATE orders
SET seller = $2
WHERE seller_id = $1;

-- name: GetOrdersByCustomerID :many
SELECT * FROM orders
WHERE customer_id = $1
ORDER BY created_at DESC;
