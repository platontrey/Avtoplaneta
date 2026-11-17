-- 1. Синхронизируем номер детали со значением первичного ключа id
UPDATE parts
SET number = id::text
WHERE number IS DISTINCT FROM id::text;

-- 2. Актуализируем последовательность part_number_seq
DO $$
DECLARE
    max_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) INTO max_id FROM parts;
    IF max_id > 0 THEN
        PERFORM setval('part_number_seq', max_id);
    END IF;
END $$;

-- 3. Создаем индекс по номеру детали для быстрых выборок
CREATE INDEX IF NOT EXISTS idx_parts_number ON parts (number);

-- 4. Триггер для автоматического присвоения номера детали из id при создании или обновлении
CREATE OR REPLACE FUNCTION set_part_number_from_id()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.id IS NULL OR NEW.id = 0 THEN
            NEW.id := nextval('parts_id_seq');
        END IF;
    END IF;
    -- Номер детали всегда равен ID детали в базе данных
    NEW.number := NEW.id::text;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_parts_sync_number ON parts;
CREATE TRIGGER trg_parts_sync_number
BEFORE INSERT OR UPDATE ON parts
FOR EACH ROW
EXECUTE FUNCTION set_part_number_from_id();
