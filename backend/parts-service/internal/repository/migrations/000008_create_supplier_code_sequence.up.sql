CREATE SEQUENCE IF NOT EXISTS defect_batch_seq START WITH 1 INCREMENT BY 1;

DO $$
DECLARE
    max_val BIGINT;
BEGIN
    SELECT COALESCE(MAX(CAST(supplier_code AS BIGINT)), 0)
    INTO max_val
    FROM parts
    WHERE supplier_code ~ '^[0-9]+$' AND LENGTH(supplier_code) < 10;

    IF max_val > 0 THEN
        PERFORM setval('defect_batch_seq', max_val);
    END IF;
END $$;
