\set ON_ERROR_STOP on

BEGIN;

CREATE TEMP TABLE jshopping_photos_stage (
    product_id BIGINT,
    photo_url TEXT,
    ordering INTEGER,
    source_file TEXT,
    size_bytes BIGINT
) ON COMMIT DROP;

\copy jshopping_photos_stage FROM '/tmp/joomshopping_photos.csv' WITH (FORMAT csv, HEADER true, ENCODING 'UTF8')

UPDATE parts p
SET photos = '[]'::jsonb,
    updated_at = NOW()
FROM (SELECT DISTINCT product_id FROM jshopping_photos_stage) source_products
WHERE p.id = source_products.product_id;

UPDATE parts p
SET photos = grouped.photos,
    updated_at = NOW()
FROM (
    SELECT product_id, jsonb_agg(photo_url ORDER BY ordering, photo_url) AS photos
    FROM (
        SELECT DISTINCT ON (product_id, photo_url)
            product_id, photo_url, ordering
        FROM jshopping_photos_stage
        WHERE photo_url <> ''
        ORDER BY product_id, photo_url, ordering
    ) existing_photos
    GROUP BY product_id
) grouped
WHERE p.id = grouped.product_id;

COMMIT;
