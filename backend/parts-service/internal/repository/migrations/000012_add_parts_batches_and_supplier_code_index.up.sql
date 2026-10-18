-- 1. Функциональный индекс для BTRIM(supplier_code), чтобы исключить Seq Scan при поиске и удалении по коду поставщика
CREATE INDEX IF NOT EXISTS idx_parts_supplier_code_btrim ON parts (BTRIM(supplier_code))
WHERE supplier_code IS NOT NULL AND BTRIM(supplier_code) != '';

-- 2. Высокопроизводительный частичный индекс для дефектных партий (quantity = 0)
CREATE INDEX IF NOT EXISTS idx_parts_zero_batches ON parts (created_at DESC, supplier_code, brand, model, car_release_date, vin)
WHERE quantity = 0 AND to_delete_at IS NULL AND deleted_at IS NULL;
