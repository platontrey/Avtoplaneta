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

-- name: GetAllParts :many
SELECT * FROM parts WHERE deleted_at IS NULL ORDER BY id;

-- name: GetPartsForXML :many
SELECT * FROM parts WHERE to_delete_at IS NULL AND quantity >= 0 AND deleted_at IS NULL ORDER BY id;

-- name: GetLastCreatedPart :one
SELECT * FROM parts WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 1;

-- name: UpdatePartPhotos :exec
UPDATE parts SET photos = $2, updated_at = NOW() WHERE id = $1;

-- name: RenameSellerInParts :many
UPDATE parts
SET salesman = $2, updated_at = NOW()
WHERE seller_id = $1 AND deleted_at IS NULL AND salesman IS DISTINCT FROM $2
RETURNING id;
