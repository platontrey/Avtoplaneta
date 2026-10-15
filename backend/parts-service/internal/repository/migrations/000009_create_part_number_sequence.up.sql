CREATE SEQUENCE IF NOT EXISTS part_number_seq START WITH 1 INCREMENT BY 1;

DO $$
DECLARE
    max_val BIGINT;
BEGIN
    SELECT COALESCE(MAX(CAST(number AS BIGINT)), 0)
    INTO max_val
    FROM parts
    WHERE number ~ '^[0-9]+$' AND LENGTH(number) < 15;

    IF max_val > 0 THEN
        PERFORM setval('part_number_seq', max_val);
    END IF;
END $$;
