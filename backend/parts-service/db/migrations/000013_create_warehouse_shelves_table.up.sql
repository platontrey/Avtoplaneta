-- 1. Справочник полок склада (любая строка без ограничений)
CREATE TABLE IF NOT EXISTS warehouse_shelves (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Индекс для быстрой сортировки и предотвращения дубликатов с разным регистром/пробелами
CREATE UNIQUE INDEX IF NOT EXISTS idx_warehouse_shelves_name_btrim
    ON warehouse_shelves (LOWER(BTRIM(name)));
