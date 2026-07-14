-- name: CreateOrder :one
INSERT INTO orders (customer_id, seller_id, seller, part, part_id, location, buyer_number, status, status_text, auto_deleted, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, part_id, quantity, price)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1;

-- name: GetOrderItemsByOrderID :many
SELECT * FROM order_items
WHERE order_id = $1;

-- name: FindAllOrders :many
SELECT * FROM orders
WHERE auto_deleted = FALSE
ORDER BY created_at DESC;

-- name: GetOrderItemsByOrderIDs :many
SELECT * FROM order_items
WHERE order_id = ANY($1::bigint[]);

-- name: FindOrderItem :one
SELECT * FROM order_items
WHERE order_id = $1 AND part_id = $2;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2
WHERE id = $1;

-- name: UpdateOrderItem :exec
UPDATE order_items
SET quantity = $2,
    price = $3
WHERE id = $1;

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
SELECT TO_CHAR(orders.created_at, 'YYYY-MM')::text as month,
       COALESCE(SUM(order_items.price * order_items.quantity), 0)::double precision as sales
FROM order_items
JOIN orders ON order_items.order_id = orders.id
WHERE orders.status = 'green' AND orders.auto_deleted = TRUE
GROUP BY TO_CHAR(orders.created_at, 'YYYY-MM')
ORDER BY month DESC;

