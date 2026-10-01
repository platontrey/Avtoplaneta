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

-- name: CreateCustomer :one
INSERT INTO customers (
    name, phone, city, preferred_tk, passport_or_inn,
    category, discount_percent, notes, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetCustomerByID :one
SELECT * FROM customers
WHERE id = $1;

-- name: GetCustomerByPhone :one
SELECT * FROM customers
WHERE phone = $1
LIMIT 1;

-- name: ListCustomersWithStats :many
SELECT c.id, c.name, c.phone, c.city, c.preferred_tk, c.passport_or_inn,
       c.category, c.discount_percent, c.notes, c.created_at, c.updated_at,
       COUNT(o.id)::bigint AS total_orders,
       COALESCE(SUM(CASE WHEN o.auto_deleted = TRUE OR o.status = 'green' THEN order_totals.total ELSE 0 END), 0)::double precision AS total_spent,
       MAX(o.created_at)::timestamptz AS last_order_at
FROM customers c
LEFT JOIN orders o ON o.customer_id = c.id
LEFT JOIN (
    SELECT order_id, SUM(price * quantity) AS total
    FROM order_items
    GROUP BY order_id
) order_totals ON order_totals.order_id = o.id
WHERE (sqlc.arg(category)::text = '' OR c.category = sqlc.arg(category)::text)
  AND (sqlc.arg(search)::text = '' OR 
       c.name ILIKE '%' || sqlc.arg(search)::text || '%' OR 
       c.phone ILIKE '%' || sqlc.arg(search)::text || '%' OR 
       c.city ILIKE '%' || sqlc.arg(search)::text || '%' OR 
       c.preferred_tk ILIKE '%' || sqlc.arg(search)::text || '%')
GROUP BY c.id
ORDER BY c.updated_at DESC
LIMIT sqlc.arg(limit_val)::int OFFSET sqlc.arg(offset_val)::int;

-- name: UpdateCustomer :one
UPDATE customers
SET name = $2,
    phone = $3,
    city = $4,
    preferred_tk = $5,
    passport_or_inn = $6,
    category = $7,
    discount_percent = $8,
    notes = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCustomer :exec
DELETE FROM customers
WHERE id = $1;

-- name: GetOrdersByCustomerID :many
SELECT * FROM orders
WHERE customer_id = $1
ORDER BY created_at DESC;

