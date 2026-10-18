-- 1. GIN индекс для мгновенной фильтрации запчастей по фото/тегам
CREATE INDEX IF NOT EXISTS idx_parts_photos_gin ON parts USING GIN (photos);

-- 2. Защита доменных инвариантов склада и цен на уровне PostgreSQL
ALTER TABLE parts ADD CONSTRAINT chk_parts_quantity CHECK (quantity >= -1);
ALTER TABLE parts ADD CONSTRAINT chk_parts_price CHECK (price >= 0);
