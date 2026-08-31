<?php

if ($argc !== 2) {
    fwrite(STDERR, "Usage: php audit_joomshopping_addresses.php <configuration.php>\n");
    exit(2);
}

require_once $argv[1];
$config = new JConfig();
$prefix = $config->dbprefix;
$db = new mysqli($config->host, $config->user, $config->password, $config->db);
if ($db->connect_errno) {
    fwrite(STDERR, "MySQL connection failed: {$db->connect_error}\n");
    exit(1);
}
$db->set_charset('utf8mb4');

function printQuery($db, $label, $sql)
{
    echo "\n== {$label} ==\n";
    $result = $db->query($sql);
    if ($result === false) {
        fwrite(STDERR, "Query failed: {$db->error}\n");
        exit(1);
    }
    while ($row = $result->fetch_assoc()) {
        echo implode("\t", array_values($row)) . "\n";
    }
    $result->free();
}

printQuery(
    $db,
    'field 44 definition',
    "SELECT id, `name_ru-RU`, type, cats, allcats FROM `{$prefix}jshopping_products_extra_fields` WHERE id = 44"
);

printQuery(
    $db,
    'all currently defined values for field 44',
    "SELECT id, field_id, `name_ru-RU` FROM `{$prefix}jshopping_products_extra_field_values` WHERE field_id = 44 ORDER BY id"
);

printQuery(
    $db,
    'numeric address ids searched across all value fields',
    "SELECT id, field_id, `name_ru-RU` FROM `{$prefix}jshopping_products_extra_field_values` WHERE id IN (878, 880, 1239, 1336) ORDER BY id"
);

printQuery(
    $db,
    'published raw address distribution',
    "SELECT extra_field_44, COUNT(*) AS products FROM `{$prefix}jshopping_products` WHERE product_publish = 1 GROUP BY extra_field_44 ORDER BY products DESC"
);

printQuery(
    $db,
    'address by supplier',
    "SELECT extra_field_49, extra_field_44, COUNT(*) AS products FROM `{$prefix}jshopping_products` WHERE product_publish = 1 GROUP BY extra_field_49, extra_field_44 ORDER BY extra_field_49, products DESC"
);

printQuery(
    $db,
    'numeric address by shelf prefix',
    "SELECT extra_field_44, SUBSTRING_INDEX(extra_field_62, '-', 1) AS shelf_prefix, COUNT(*) AS products FROM `{$prefix}jshopping_products` WHERE product_publish = 1 AND extra_field_44 IN ('878','880','1239','1336') GROUP BY extra_field_44, shelf_prefix ORDER BY extra_field_44, products DESC LIMIT 100"
);

printQuery(
    $db,
    'address by first shelf letter',
    "SELECT extra_field_44, LEFT(TRIM(extra_field_62), 1) AS shelf_initial, COUNT(*) AS products FROM `{$prefix}jshopping_products` WHERE product_publish = 1 GROUP BY extra_field_44, shelf_initial ORDER BY extra_field_44, products DESC"
);

printQuery(
    $db,
    'address date ranges',
    "SELECT extra_field_44, MIN(product_date_added) AS oldest, MAX(product_date_added) AS newest, COUNT(*) AS products FROM `{$prefix}jshopping_products` WHERE product_publish = 1 GROUP BY extra_field_44 ORDER BY products DESC"
);

printQuery(
    $db,
    'value ids neighboring deleted address ids',
    "SELECT id, field_id, `name_ru-RU` FROM `{$prefix}jshopping_products_extra_field_values` WHERE (id BETWEEN 870 AND 890) OR (id BETWEEN 1230 AND 1245) OR (id BETWEEN 1325 AND 1345) ORDER BY id"
);

printQuery(
    $db,
    'small deleted address samples with resolved shelves',
    "SELECT p.product_id, p.`name_ru-RU`, p.extra_field_44, COALESCE(NULLIF(TRIM(p.extra_field_62), ''), NULLIF(TRIM(v.`name_ru-RU`), '')) AS shelf, p.product_date_added FROM `{$prefix}jshopping_products` p LEFT JOIN `{$prefix}jshopping_products_extra_field_values` v ON v.id = CAST(p.extra_field_61 AS UNSIGNED) AND v.field_id = 61 WHERE p.product_publish = 1 AND p.extra_field_44 IN ('878','880','1239') ORDER BY p.extra_field_44, p.product_id"
);

$db->close();
