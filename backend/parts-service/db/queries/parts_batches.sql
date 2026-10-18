-- name: CountZeroQuantityBySupplier :one
SELECT COUNT(*)::bigint FROM parts
WHERE supplier_code = $1 AND quantity = 0 AND to_delete_at IS NULL AND deleted_at IS NULL;

-- name: DeleteZeroQuantityBySupplier :execrows
DELETE FROM parts
WHERE supplier_code = $1 AND quantity = 0 AND to_delete_at IS NULL AND deleted_at IS NULL;

-- name: GetSupplierCodes :many
SELECT DISTINCT supplier_code FROM parts
WHERE supplier_code IS NOT NULL AND supplier_code != '' AND quantity = 0 AND to_delete_at IS NULL AND deleted_at IS NULL
ORDER BY supplier_code;

-- name: GetNextDefectBatchSequence :one
SELECT nextval('defect_batch_seq')::bigint;

-- name: GetMaxSupplierCodeFallback :one
SELECT COALESCE(MAX(CAST(supplier_code AS BIGINT)), 0) + 1 AS next_code
FROM parts
WHERE supplier_code ~ '^[0-9]+$' AND LENGTH(supplier_code) < 10;

-- name: GetNextPartNumberSequence :one
SELECT nextval('part_number_seq')::bigint;

-- name: GetMaxPartNumberFallback :one
SELECT COALESCE(MAX(CAST(number AS BIGINT)), 0) + 1 AS next_number
FROM parts
WHERE number ~ '^[0-9]+$' AND LENGTH(number) < 15;

-- name: DeleteZeroQuantityPartsBySupplierCode :many
DELETE FROM parts
WHERE BTRIM(supplier_code) = BTRIM($1)
  AND quantity = 0 AND to_delete_at IS NULL AND deleted_at IS NULL
RETURNING id;

-- name: DeleteZeroQuantityPartsByCar :many
DELETE FROM parts
WHERE (supplier_code IS NULL OR BTRIM(supplier_code) = '')
  AND COALESCE(BTRIM(brand), '') = $1
  AND COALESCE(BTRIM(model), '') = $2
  AND COALESCE(BTRIM(car_release_date), '') = $3
  AND COALESCE(BTRIM(vin), '') = $4
  AND quantity = 0 AND to_delete_at IS NULL AND deleted_at IS NULL
RETURNING id;

-- name: GetSupplierBatches :many
WITH batch_identifiers AS (
    SELECT
        CASE
            WHEN supplier_code IS NOT NULL AND BTRIM(supplier_code) != '' THEN BTRIM(supplier_code)
            ELSE 'car:' || COALESCE(BTRIM(brand), '') || '|' || COALESCE(BTRIM(model), '') || '|' || COALESCE(BTRIM(car_release_date), '') || '|' || COALESCE(BTRIM(vin), '')
        END AS batch_code,
        brand,
        model,
        car_release_date,
        vin,
        created_at
    FROM parts
    WHERE quantity = 0 AND to_delete_at IS NULL AND deleted_at IS NULL
)
SELECT
    batch_code,
    COALESCE(MAX(NULLIF(BTRIM(brand), '')), '')::text AS brand_name,
    COALESCE(MAX(NULLIF(BTRIM(model), '')), '')::text AS model_name,
    COALESCE(MAX(NULLIF(BTRIM(car_release_date), '')), '')::text AS release_date,
    COALESCE(MAX(NULLIF(BTRIM(vin), '')), '')::text AS car_vin,
    COUNT(*)::bigint AS zero_count
FROM batch_identifiers
GROUP BY batch_code
ORDER BY MAX(created_at) DESC NULLS LAST, batch_code DESC;
