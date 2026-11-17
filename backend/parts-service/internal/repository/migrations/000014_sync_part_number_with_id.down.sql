DROP TRIGGER IF EXISTS trg_parts_sync_number ON parts;
DROP FUNCTION IF EXISTS set_part_number_from_id();
DROP INDEX IF EXISTS idx_parts_number;
