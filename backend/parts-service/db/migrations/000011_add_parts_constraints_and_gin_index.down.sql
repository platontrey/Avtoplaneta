ALTER TABLE parts DROP CONSTRAINT IF EXISTS chk_parts_price;
ALTER TABLE parts DROP CONSTRAINT IF EXISTS chk_parts_quantity;
DROP INDEX IF EXISTS idx_parts_photos_gin;
