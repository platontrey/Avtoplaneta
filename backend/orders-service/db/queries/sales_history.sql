-- name: CreateSalesHistory :one
INSERT INTO sales_history (month, sales, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMonthlySales :many
WITH completed_orders AS (
    SELECT 
        TO_CHAR(orders.created_at, 'YYYY-MM')::text AS month,
        GREATEST(COALESCE(SUM(order_items.price * order_items.quantity), 0) - orders.discount, 0)::double precision AS order_total
    FROM orders
    JOIN order_items ON order_items.order_id = orders.id
    WHERE orders.status IN ('green', 'completed') OR orders.payment_status = 'paid' OR orders.auto_deleted = TRUE
    GROUP BY orders.id, orders.created_at, orders.discount
)
SELECT month, COALESCE(SUM(order_total), 0)::double precision AS sales
FROM completed_orders
GROUP BY month
ORDER BY month DESC;
