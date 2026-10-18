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
WITH order_totals AS (
    SELECT order_id, SUM(price * quantity) AS total
    FROM order_items
    GROUP BY order_id
)
SELECT c.id, c.name, c.phone, c.city, c.preferred_tk, c.passport_or_inn,
       c.category, c.discount_percent, c.notes, c.created_at, c.updated_at,
       COUNT(o.id)::bigint AS total_orders,
       COALESCE(SUM(CASE WHEN o.auto_deleted = TRUE OR o.status = 'green' THEN order_totals.total ELSE 0 END), 0)::double precision AS total_spent,
       MAX(o.created_at)::timestamptz AS last_order_at
FROM customers c
LEFT JOIN orders o ON o.customer_id = c.id
LEFT JOIN order_totals ON order_totals.order_id = o.id
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
