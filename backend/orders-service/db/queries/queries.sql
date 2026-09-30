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

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, part_id, quantity, price, part_name_snapshot)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1;

-- name: GetOrderItemsByOrderID :many
SELECT * FROM order_items
WHERE order_id = $1
ORDER BY id ASC;

-- name: FindAllOrders :many
SELECT * FROM orders
WHERE auto_deleted = FALSE
ORDER BY created_at DESC;

-- name: FindCompletedOrders :many
SELECT * FROM orders
WHERE auto_deleted = TRUE
ORDER BY COALESCE(completed_at, created_at) DESC
LIMIT 200;

-- name: GetOrderItemsByOrderIDs :many
SELECT * FROM order_items
WHERE order_id = ANY($1::bigint[])
ORDER BY id ASC;

-- name: FindOrderItem :one
SELECT * FROM order_items
WHERE order_id = $1 AND part_id = $2;

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

-- name: UpdateOrderItem :exec
UPDATE order_items
SET quantity = $2,
    price = $3
WHERE id = $1;

-- name: DeleteOrderItem :exec
DELETE FROM order_items
WHERE id = $1 AND order_id = $2;

-- name: DeleteOrder :exec
DELETE FROM orders
WHERE id = $1;

-- name: DeleteOrderItemsByOrderID :exec
DELETE FROM order_items
WHERE order_id = $1;

-- name: MarkExpiredAsAutoDeleted :exec
UPDATE orders
SET auto_deleted = TRUE
WHERE created_at <= $1 AND auto_deleted = FALSE;

-- name: CreateSalesHistory :one
INSERT INTO sales_history (month, sales, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMonthlySales :many
SELECT month, COALESCE(SUM(order_total), 0)::double precision AS sales
FROM (
    SELECT TO_CHAR(COALESCE(orders.completed_at, orders.created_at), 'YYYY-MM')::text AS month,
           GREATEST(COALESCE(SUM(order_items.price * order_items.quantity), 0) - orders.discount, 0)::double precision AS order_total
    FROM orders
    JOIN order_items ON order_items.order_id = orders.id
    WHERE orders.status = 'green' AND orders.auto_deleted = TRUE
    GROUP BY orders.id, orders.completed_at, orders.created_at, orders.discount
) AS completed_orders
GROUP BY month
ORDER BY month DESC;

-- name: UpdateSellerName :exec
UPDATE orders
SET seller = $2
WHERE seller_id = $1;
