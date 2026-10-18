-- name: CreatePart :one
INSERT INTO parts (
    name, quantity, description, category, price, salesman, location, address, status,
    brand, model, photos, seller_id, to_delete_at, vin,
    body_brand, engine_brand, car_release_date, car_release_period, front_rear, left_right, top_bottom,
    number, manufacturer, manufacturer_code, oem_code, color, condition,
    supplier_code, defect, transmission, transmission_model, drive, wear_percentage,
    season, diameter, width, profile, tire_quantity, drilling, "offset",
    center_hole_diameter, tire_model, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19, $20, $21, $22,
    $23, $24, $25, $26, $27, $28,
    $29, $30, $31, $32, $33, $34,
    $35, $36, $37, $38, $39, $40, $41,
    $42, $43, NOW(), NOW()
) RETURNING *;

-- name: GetPartByID :one
SELECT * FROM parts WHERE id = $1 AND deleted_at IS NULL;

-- name: DeletePart :exec
DELETE FROM parts WHERE id = $1;

-- name: SoftDeletePart :exec
UPDATE parts SET deleted_at = NOW() WHERE id = $1;

-- name: MarkForDeletion :exec
UPDATE parts SET to_delete_at = $2, updated_at = NOW() WHERE id = $1;

-- name: DeleteExpiredParts :execrows
DELETE FROM parts WHERE to_delete_at IS NOT NULL AND to_delete_at <= $1;

-- name: BulkDeleteByIDs :exec
DELETE FROM parts WHERE id = ANY($1::bigint[]);

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

-- name: GetAllParts :many
SELECT * FROM parts WHERE deleted_at IS NULL ORDER BY id;

-- name: GetPartsForXML :many
SELECT * FROM parts WHERE to_delete_at IS NULL AND quantity >= 0 AND deleted_at IS NULL ORDER BY id;

-- name: GetLastCreatedPart :one
SELECT * FROM parts WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 1;

-- name: UpdatePartPhotos :exec
UPDATE parts SET photos = $2, updated_at = NOW() WHERE id = $1;

-- name: GetStatsTotals :one
SELECT
    COUNT(*)::bigint AS total_parts,
    COALESCE(SUM(quantity), 0)::bigint AS total_quantity,
    COALESCE(SUM(price * quantity), 0)::float8 AS total_value
FROM parts
WHERE to_delete_at IS NULL AND quantity >= 1 AND deleted_at IS NULL;

-- name: GetStatsTotalsBefore :one
SELECT
    COUNT(*)::bigint AS prev_parts,
    COALESCE(SUM(quantity), 0)::bigint AS prev_quantity,
    COALESCE(SUM(price * quantity), 0)::float8 AS prev_value
FROM parts
WHERE to_delete_at IS NULL
  AND quantity >= 1
  AND created_at < $1
  AND (deleted_at IS NULL OR deleted_at >= $1);

-- name: GetStatsCategories :many
SELECT category AS name, COUNT(*)::bigint AS count
FROM parts
WHERE to_delete_at IS NULL AND quantity >= 1 AND deleted_at IS NULL AND category != ''
GROUP BY category
ORDER BY count DESC;

-- name: DecreasePartQuantity :exec
UPDATE parts
SET quantity = CASE WHEN quantity - sqlc.arg(amount)::integer <= 0 THEN -1 ELSE quantity - sqlc.arg(amount)::integer END
WHERE id = sqlc.arg(id);

-- name: IncreasePartQuantity :exec
UPDATE parts
SET quantity = CASE WHEN quantity = -1 THEN sqlc.arg(amount)::integer ELSE quantity + sqlc.arg(amount)::integer END
WHERE id = sqlc.arg(id);

-- name: RecordStockOperation :execrows
INSERT INTO part_stock_operations (operation_id, part_id, operation_type, amount)
VALUES ($1, $2, $3, $4)
ON CONFLICT (operation_id, part_id, operation_type) DO NOTHING;

-- name: RenameSellerInParts :many
UPDATE parts
SET salesman = $2, updated_at = NOW()
WHERE seller_id = $1 AND deleted_at IS NULL AND salesman IS DISTINCT FROM $2
RETURNING id;

-- name: GetInventoryVersion :one
SELECT MAX(updated_at)::timestamptz AS last_change, COUNT(*)::bigint AS alive
FROM parts
WHERE deleted_at IS NULL;

-- name: BulkUpdateFieldsByIDs :many
UPDATE parts
SET
    price = CASE
        WHEN sqlc.arg(has_price)::boolean AND sqlc.arg(price_mode)::text = 'percent'
            THEN ROUND((price * (1.0 + sqlc.arg(price_val)::float8 / 100.0))::numeric, 2)::double precision
        WHEN sqlc.arg(has_price)::boolean AND sqlc.arg(price_mode)::text = 'offset'
            THEN GREATEST(0.0, price + sqlc.arg(price_val)::float8)
        WHEN sqlc.arg(has_price)::boolean AND sqlc.arg(price_mode)::text = 'set'
            THEN sqlc.arg(price_val)::float8
        ELSE price
    END,
    category = CASE WHEN sqlc.arg(has_category)::boolean THEN sqlc.arg(category)::text ELSE category END,
    location = CASE WHEN sqlc.arg(has_location)::boolean THEN sqlc.arg(location)::text ELSE location END,
    address = CASE WHEN sqlc.arg(has_address)::boolean THEN sqlc.arg(address)::text ELSE address END,
    salesman = CASE WHEN sqlc.arg(has_salesman)::boolean THEN sqlc.arg(salesman)::text ELSE salesman END,
    status = CASE WHEN sqlc.arg(has_status)::boolean THEN sqlc.arg(status)::boolean ELSE status END,
    body_brand = CASE WHEN sqlc.arg(has_body_brand)::boolean THEN sqlc.arg(body_brand)::text ELSE body_brand END,
    engine_brand = CASE WHEN sqlc.arg(has_engine_brand)::boolean THEN sqlc.arg(engine_brand)::text ELSE engine_brand END,
    updated_at = NOW()
WHERE id = ANY(sqlc.arg(ids)::bigint[]) AND deleted_at IS NULL AND quantity >= 0
RETURNING *;

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
SELECT
    sub.batch_code,
    COALESCE(MAX(NULLIF(BTRIM(sub.brand), '')), '')::text AS brand_name,
    COALESCE(MAX(NULLIF(BTRIM(sub.model), '')), '')::text AS model_name,
    COALESCE(MAX(NULLIF(BTRIM(sub.car_release_date), '')), '')::text AS release_date,
    COALESCE(MAX(NULLIF(BTRIM(sub.vin), '')), '')::text AS car_vin,
    COUNT(*)::bigint AS zero_count
FROM (
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
) sub
GROUP BY sub.batch_code
ORDER BY MAX(sub.created_at) DESC NULLS LAST, sub.batch_code DESC;
