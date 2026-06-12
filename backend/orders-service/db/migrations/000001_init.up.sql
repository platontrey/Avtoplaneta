-- 000001_init.up.sql
CREATE TABLE IF NOT EXISTS parts (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL DEFAULT 0,
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT '',
    price DOUBLE PRECISION NOT NULL DEFAULT 0,
    salesman TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',
    status BOOLEAN NOT NULL DEFAULT false,
    brand TEXT NOT NULL DEFAULT '',
    model TEXT NOT NULL DEFAULT '',
    photos JSONB DEFAULT '[]'::jsonb,
    seller_id BIGINT NOT NULL DEFAULT 0,
    to_delete_at TIMESTAMPTZ DEFAULT NULL,
    vin TEXT NOT NULL DEFAULT '',
    body_brand TEXT NOT NULL DEFAULT '',
    engine_brand TEXT NOT NULL DEFAULT '',
    car_release_date TEXT NOT NULL DEFAULT '',
    front_rear TEXT NOT NULL DEFAULT '',
    left_right TEXT NOT NULL DEFAULT '',
    top_bottom TEXT NOT NULL DEFAULT '',
    number TEXT NOT NULL DEFAULT '',
    manufacturer TEXT NOT NULL DEFAULT '',
    manufacturer_code TEXT NOT NULL DEFAULT '',
    oem_code TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '',
    condition TEXT NOT NULL DEFAULT '',
    supplier_code TEXT NOT NULL DEFAULT '',
    defect TEXT NOT NULL DEFAULT '',
    transmission TEXT NOT NULL DEFAULT '',
    drive TEXT NOT NULL DEFAULT '',
    wear_percentage TEXT NOT NULL DEFAULT '',
    season TEXT NOT NULL DEFAULT '',
    diameter TEXT NOT NULL DEFAULT '',
    width TEXT NOT NULL DEFAULT '',
    profile TEXT NOT NULL DEFAULT '',
    tire_quantity TEXT NOT NULL DEFAULT '',
    drilling TEXT NOT NULL DEFAULT '',
    "offset" TEXT NOT NULL DEFAULT '',
    center_hole_diameter TEXT NOT NULL DEFAULT '',
    tire_model TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    seller VARCHAR(255) NOT NULL DEFAULT '',
    part VARCHAR(255) NOT NULL DEFAULT '',
    part_id BIGINT NOT NULL,
    location VARCHAR(255) NOT NULL DEFAULT '',
    buyer_number VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'red',
    status_text VARCHAR(255) NOT NULL DEFAULT '',
    auto_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    part_id BIGINT NOT NULL,
    quantity INTEGER NOT NULL,
    price DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS sales_history (
    id BIGSERIAL PRIMARY KEY,
    month VARCHAR(7) NOT NULL,
    sales DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_part_id ON order_items (part_id);

CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_auto_deleted ON orders (auto_deleted);

-- Remove buggy GORM index if it exists
DROP INDEX IF EXISTS idx_orders_user_id;

-- Index for parts
CREATE INDEX IF NOT EXISTS idx_parts_id_orders ON parts (id);
CREATE INDEX IF NOT EXISTS idx_parts_quantity_orders ON parts (quantity);

-- Index for sales_history
CREATE INDEX IF NOT EXISTS idx_sales_history_created_at ON sales_history (created_at DESC);
