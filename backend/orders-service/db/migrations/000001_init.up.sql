-- 000001_init.up.sql

CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    part_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    auto_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(50) NOT NULL DEFAULT 'red',
    seller VARCHAR(255) NOT NULL DEFAULT '',
    part VARCHAR(255) NOT NULL DEFAULT '',
    location VARCHAR(255) NOT NULL DEFAULT '',
    buyer_number VARCHAR(255) NOT NULL DEFAULT '',
    status_text VARCHAR(255) NOT NULL DEFAULT ''
) WITH (fillfactor = 90);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    part_id BIGINT NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    quantity INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS sales_history (
    id BIGSERIAL PRIMARY KEY,
    sales DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    month VARCHAR(7) NOT NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_part_id ON order_items (part_id);

CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_auto_deleted ON orders (auto_deleted);

-- Remove buggy GORM index if it exists
DROP INDEX IF EXISTS idx_orders_user_id;

-- Index for sales_history
CREATE INDEX IF NOT EXISTS idx_sales_history_created_at ON sales_history (created_at DESC);
