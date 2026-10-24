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
    name = CASE WHEN sqlc.arg(has_name)::boolean THEN sqlc.arg(name)::text ELSE name END,
    quantity = CASE WHEN sqlc.arg(has_quantity)::boolean THEN sqlc.arg(quantity)::int4 ELSE quantity END,
    description = CASE WHEN sqlc.arg(has_description)::boolean THEN sqlc.arg(description)::text ELSE description END,
    category = CASE WHEN sqlc.arg(has_category)::boolean THEN sqlc.arg(category)::text ELSE category END,
    location = CASE WHEN sqlc.arg(has_location)::boolean THEN sqlc.arg(location)::text ELSE location END,
    address = CASE WHEN sqlc.arg(has_address)::boolean THEN sqlc.arg(address)::text ELSE address END,
    salesman = CASE WHEN sqlc.arg(has_salesman)::boolean THEN sqlc.arg(salesman)::text ELSE salesman END,
    seller_id = CASE WHEN sqlc.arg(has_seller_id)::boolean THEN sqlc.arg(seller_id)::bigint ELSE seller_id END,
    status = CASE WHEN sqlc.arg(has_status)::boolean THEN sqlc.arg(status)::boolean ELSE status END,
    brand = CASE WHEN sqlc.arg(has_brand)::boolean THEN sqlc.arg(brand)::text ELSE brand END,
    model = CASE WHEN sqlc.arg(has_model)::boolean THEN sqlc.arg(model)::text ELSE model END,
    vin = CASE WHEN sqlc.arg(has_vin)::boolean THEN sqlc.arg(vin)::text ELSE vin END,
    body_brand = CASE WHEN sqlc.arg(has_body_brand)::boolean THEN sqlc.arg(body_brand)::text ELSE body_brand END,
    engine_brand = CASE WHEN sqlc.arg(has_engine_brand)::boolean THEN sqlc.arg(engine_brand)::text ELSE engine_brand END,
    car_release_date = CASE WHEN sqlc.arg(has_car_release_date)::boolean THEN sqlc.arg(car_release_date)::text ELSE car_release_date END,
    car_release_period = CASE WHEN sqlc.arg(has_car_release_period)::boolean THEN sqlc.arg(car_release_period)::text ELSE car_release_period END,
    front_rear = CASE WHEN sqlc.arg(has_front_rear)::boolean THEN sqlc.arg(front_rear)::text ELSE front_rear END,
    left_right = CASE WHEN sqlc.arg(has_left_right)::boolean THEN sqlc.arg(left_right)::text ELSE left_right END,
    top_bottom = CASE WHEN sqlc.arg(has_top_bottom)::boolean THEN sqlc.arg(top_bottom)::text ELSE top_bottom END,
    number = CASE WHEN sqlc.arg(has_number)::boolean THEN sqlc.arg(number)::text ELSE number END,
    manufacturer = CASE WHEN sqlc.arg(has_manufacturer)::boolean THEN sqlc.arg(manufacturer)::text ELSE manufacturer END,
    manufacturer_code = CASE WHEN sqlc.arg(has_manufacturer_code)::boolean THEN sqlc.arg(manufacturer_code)::text ELSE manufacturer_code END,
    oem_code = CASE WHEN sqlc.arg(has_oem_code)::boolean THEN sqlc.arg(oem_code)::text ELSE oem_code END,
    color = CASE WHEN sqlc.arg(has_color)::boolean THEN sqlc.arg(color)::text ELSE color END,
    condition = CASE WHEN sqlc.arg(has_condition)::boolean THEN sqlc.arg(condition)::text ELSE condition END,
    supplier_code = CASE WHEN sqlc.arg(has_supplier_code)::boolean THEN sqlc.arg(supplier_code)::text ELSE supplier_code END,
    defect = CASE WHEN sqlc.arg(has_defect)::boolean THEN sqlc.arg(defect)::text ELSE defect END,
    transmission = CASE WHEN sqlc.arg(has_transmission)::boolean THEN sqlc.arg(transmission)::text ELSE transmission END,
    transmission_model = CASE WHEN sqlc.arg(has_transmission_model)::boolean THEN sqlc.arg(transmission_model)::text ELSE transmission_model END,
    drive = CASE WHEN sqlc.arg(has_drive)::boolean THEN sqlc.arg(drive)::text ELSE drive END,
    wear_percentage = CASE WHEN sqlc.arg(has_wear_percentage)::boolean THEN sqlc.arg(wear_percentage)::text ELSE wear_percentage END,
    season = CASE WHEN sqlc.arg(has_season)::boolean THEN sqlc.arg(season)::text ELSE season END,
    diameter = CASE WHEN sqlc.arg(has_diameter)::boolean THEN sqlc.arg(diameter)::text ELSE diameter END,
    width = CASE WHEN sqlc.arg(has_width)::boolean THEN sqlc.arg(width)::text ELSE width END,
    profile = CASE WHEN sqlc.arg(has_profile)::boolean THEN sqlc.arg(profile)::text ELSE profile END,
    tire_quantity = CASE WHEN sqlc.arg(has_tire_quantity)::boolean THEN sqlc.arg(tire_quantity)::text ELSE tire_quantity END,
    drilling = CASE WHEN sqlc.arg(has_drilling)::boolean THEN sqlc.arg(drilling)::text ELSE drilling END,
    "offset" = CASE WHEN sqlc.arg(has_offset)::boolean THEN sqlc.arg(offset_val)::text ELSE "offset" END,
    center_hole_diameter = CASE WHEN sqlc.arg(has_center_hole_diameter)::boolean THEN sqlc.arg(center_hole_diameter)::text ELSE center_hole_diameter END,
    tire_model = CASE WHEN sqlc.arg(has_tire_model)::boolean THEN sqlc.arg(tire_model)::text ELSE tire_model END,
    updated_at = NOW()
WHERE id = ANY(sqlc.arg(ids)::bigint[]) AND deleted_at IS NULL AND quantity >= 0
RETURNING *;
