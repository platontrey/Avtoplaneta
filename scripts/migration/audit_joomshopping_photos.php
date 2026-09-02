<?php

if ($argc !== 3) {
    fwrite(STDERR, "Usage: php audit_joomshopping_photos.php <configuration.php> <img_products_dir>\n");
    exit(2);
}

require_once $argv[1];
$imageDir = rtrim($argv[2], '/');
$config = new JConfig();
$prefix = $config->dbprefix;
$db = new mysqli($config->host, $config->user, $config->password, $config->db);
if ($db->connect_errno) {
    fwrite(STDERR, "MySQL connection failed: {$db->connect_error}\n");
    exit(1);
}
$db->set_charset('utf8mb4');

function scalarQuery($db, $sql)
{
    $result = $db->query($sql);
    if ($result === false) {
        fwrite(STDERR, "Query failed: {$db->error}\n");
        exit(1);
    }
    $row = $result->fetch_row();
    $result->free();
    return $row[0];
}

echo "== schema ==\n";
$schema = $db->query("SHOW COLUMNS FROM `{$prefix}jshopping_products_images`");
while ($row = $schema->fetch_assoc()) {
    echo $row['Field'] . "\t" . $row['Type'] . "\n";
}
$schema->free();

$join = " FROM `{$prefix}jshopping_products_images` i " .
    "JOIN `{$prefix}jshopping_products` p ON p.product_id = i.product_id " .
    "WHERE p.product_publish = 1 AND TRIM(i.image_name) <> ''";

echo "== counts ==\n";
echo "published_images\t" . scalarQuery($db, "SELECT COUNT(*)" . $join) . "\n";
echo "published_products_with_images\t" . scalarQuery($db, "SELECT COUNT(DISTINCT i.product_id)" . $join) . "\n";
echo "distinct_image_names\t" . scalarQuery($db, "SELECT COUNT(DISTINCT i.image_name)" . $join) . "\n";

echo "== samples ==\n";
$samples = $db->query(
    "SELECT i.image_id, i.product_id, i.image_name, i.ordering" . $join .
    " ORDER BY i.product_id, i.ordering, i.image_id LIMIT 20"
);
while ($row = $samples->fetch_assoc()) {
    $name = basename((string) $row['image_name']);
    $variants = array($name, 'full_' . $name, 'thumb_' . $name);
    $found = array();
    foreach ($variants as $variant) {
        $path = $imageDir . '/' . $variant;
        if (is_file($path)) {
            $found[] = $variant . ':' . filesize($path);
        }
    }
    echo $row['image_id'] . "\t" . $row['product_id'] . "\t" . $name . "\t" .
        $row['ordering'] . "\t" . implode(',', $found) . "\n";
}
$samples->free();
$db->close();
