-- 000003_create_customers.up.sql

CREATE TABLE IF NOT EXISTS customers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT '',
    phone VARCHAR(64) NOT NULL DEFAULT '',
    city VARCHAR(128) NOT NULL DEFAULT '',
    preferred_tk VARCHAR(128) NOT NULL DEFAULT '',
    passport_or_inn VARCHAR(128) NOT NULL DEFAULT '',
    category VARCHAR(32) NOT NULL DEFAULT 'regular',
    discount_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers (phone);
CREATE INDEX IF NOT EXISTS idx_customers_name ON customers (name);
CREATE INDEX IF NOT EXISTS idx_customers_category ON customers (category);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders (customer_id);

-- Backfill customers from existing orders with non-empty buyer_number
INSERT INTO customers (phone, name, created_at, updated_at)
SELECT DISTINCT ON (TRIM(buyer_number))
    TRIM(buyer_number) AS phone,
    TRIM(buyer_number) AS name,
    MIN(created_at) OVER (PARTITION BY TRIM(buyer_number)) AS created_at,
    MAX(created_at) OVER (PARTITION BY TRIM(buyer_number)) AS updated_at
FROM orders
WHERE buyer_number IS NOT NULL
  AND TRIM(buyer_number) != ''
  AND TRIM(buyer_number) != 'Продажа на месте'
  AND TRIM(buyer_number) != 'Без контакта'
  AND NOT EXISTS (
      SELECT 1 FROM customers c WHERE c.phone = TRIM(buyer_number)
  )
ORDER BY TRIM(buyer_number), created_at ASC;

-- Link existing orders to newly created customers
UPDATE orders o
SET customer_id = c.id
FROM customers c
WHERE TRIM(o.buyer_number) = c.phone
  AND (o.customer_id = 0 OR o.customer_id IS NULL);
