-- Откат миграции: удаление индексов и таблицы
DROP INDEX IF EXISTS idx_items_active;
DROP INDEX IF EXISTS idx_items_created_at;
DROP INDEX IF EXISTS idx_items_status;
DROP TABLE IF EXISTS items;
