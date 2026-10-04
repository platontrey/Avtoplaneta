-- 000005_widen_order_and_customer_fields.up.sql

ALTER TABLE orders
    ALTER COLUMN order_number TYPE TEXT,
    ALTER COLUMN transport_company TYPE TEXT,
    ALTER COLUMN tracking_number TYPE TEXT,
    ALTER COLUMN part TYPE TEXT,
    ALTER COLUMN location TYPE TEXT,
    ALTER COLUMN buyer_number TYPE TEXT,
    ALTER COLUMN status_text TYPE TEXT;

ALTER TABLE customers
    ALTER COLUMN preferred_tk TYPE TEXT,
    ALTER COLUMN city TYPE TEXT,
    ALTER COLUMN passport_or_inn TYPE TEXT,
    ALTER COLUMN name TYPE TEXT;

ALTER TABLE order_items
    ALTER COLUMN part_name_snapshot TYPE TEXT;
