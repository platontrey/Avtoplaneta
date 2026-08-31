ALTER TABLE parts
    ADD COLUMN IF NOT EXISTS address TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_parts_address ON parts (address);
