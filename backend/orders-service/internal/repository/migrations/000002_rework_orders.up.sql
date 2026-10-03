-- 000002_rework_orders.up.sql

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS order_number VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS source VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payment_status VARCHAR(50) NOT NULL DEFAULT 'unpaid',
    ADD COLUMN IF NOT EXISTS warehouse_status VARCHAR(50) NOT NULL DEFAULT 'inspecting',
    ADD COLUMN IF NOT EXISTS delivery_method VARCHAR(50) NOT NULL DEFAULT 'pickup',
    ADD COLUMN IF NOT EXISTS transport_company VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS tracking_number VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS discount DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP WITH TIME ZONE NULL,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS part_name_snapshot VARCHAR(255) NOT NULL DEFAULT '';

-- Backfill completed_at for already completed orders
UPDATE orders
SET completed_at = created_at,
    payment_status = 'paid',
    warehouse_status = 'ready'
WHERE (status = 'green' OR auto_deleted = TRUE) AND completed_at IS NULL;

-- Backfill part_name_snapshot for existing items from orders.part
UPDATE order_items oi
SET part_name_snapshot = o.part
FROM orders o
WHERE oi.order_id = o.id
  AND oi.part_name_snapshot = ''
  AND o.part != '';

CREATE INDEX IF NOT EXISTS idx_orders_completed_at ON orders (completed_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_payment_status ON orders (payment_status);
CREATE INDEX IF NOT EXISTS idx_orders_warehouse_status ON orders (warehouse_status);
