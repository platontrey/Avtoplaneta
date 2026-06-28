-- name: CreatePart :one
INSERT INTO parts (
    name, quantity, description, category, price, salesman, location, status,
    brand, model, photos, seller_id, to_delete_at, vin,
    body_brand, engine_brand, car_release_date, front_rear, left_right, top_bottom,
    number, manufacturer, manufacturer_code, oem_code, color, condition,
    supplier_code, defect, transmission, drive, wear_percentage,
    season, diameter, width, profile, tire_quantity, drilling, "offset",
    center_hole_diameter, tire_model, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26,
    $27, $28, $29, $30, $31,
    $32, $33, $34, $35, $36, $37, $38,
    $39, $40, NOW(), NOW()
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

-- name: GetStatsCategories :many
SELECT category AS name, COUNT(*)::bigint AS count
FROM parts
WHERE to_delete_at IS NULL AND quantity >= 1 AND deleted_at IS NULL AND category != ''
GROUP BY category
ORDER BY count DESC;
