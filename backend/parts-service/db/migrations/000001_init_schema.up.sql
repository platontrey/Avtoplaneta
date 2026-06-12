CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS parts (
    id BIGSERIAL PRIMARY KEY,

    -- PartCore
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

    -- PartSpecifications
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

    -- PartTireSpecifications
    season TEXT NOT NULL DEFAULT '',
    diameter TEXT NOT NULL DEFAULT '',
    width TEXT NOT NULL DEFAULT '',
    profile TEXT NOT NULL DEFAULT '',
    tire_quantity TEXT NOT NULL DEFAULT '',
    drilling TEXT NOT NULL DEFAULT '',
    "offset" TEXT NOT NULL DEFAULT '',
    center_hole_diameter TEXT NOT NULL DEFAULT '',
    tire_model TEXT NOT NULL DEFAULT '',

    -- Timestamps (GORM adds these by default)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS earnings (
    id BIGSERIAL PRIMARY KEY,
    total_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- B-tree indexes
CREATE INDEX IF NOT EXISTS idx_parts_category ON parts (category);
CREATE INDEX IF NOT EXISTS idx_parts_brand ON parts (brand);
CREATE INDEX IF NOT EXISTS idx_parts_model ON parts (model);
CREATE INDEX IF NOT EXISTS idx_parts_location ON parts (location);
CREATE INDEX IF NOT EXISTS idx_parts_salesman ON parts (salesman);
CREATE INDEX IF NOT EXISTS idx_parts_status ON parts (status);
CREATE INDEX IF NOT EXISTS idx_parts_supplier_code ON parts (supplier_code);
CREATE INDEX IF NOT EXISTS idx_parts_to_delete_at ON parts (to_delete_at) WHERE to_delete_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_parts_quantity ON parts (quantity);
CREATE INDEX IF NOT EXISTS idx_parts_price ON parts (price);
CREATE INDEX IF NOT EXISTS idx_parts_created_at ON parts (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_parts_deleted_at ON parts (deleted_at);

-- GIN index for JSONB photos
CREATE INDEX IF NOT EXISTS idx_parts_photos ON parts USING GIN (photos);
