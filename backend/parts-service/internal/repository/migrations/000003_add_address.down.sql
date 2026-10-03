DROP INDEX IF EXISTS idx_parts_address;

ALTER TABLE parts
    DROP COLUMN IF EXISTS address;
