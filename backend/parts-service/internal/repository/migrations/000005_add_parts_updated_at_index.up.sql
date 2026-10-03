-- Валидатор условных запросов (ETag) для прайс-листа, статистики и справочников
-- берёт MAX(updated_at) и COUNT(*) по живым запчастям. Без индекса это полное
-- сканирование, и проверка стоила бы дороже самого ответа.
--
-- Индекс частичный: предикат совпадает с условием запроса, поэтому его хватает
-- и для MAX (скан по индексу с конца), и для COUNT (index-only scan).
CREATE INDEX IF NOT EXISTS idx_parts_updated_at_alive
    ON parts (updated_at DESC)
    WHERE deleted_at IS NULL;
