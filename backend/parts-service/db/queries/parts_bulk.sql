-- name: BulkDeleteByIDs :exec
DELETE FROM parts WHERE id = ANY($1::bigint[]);

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
