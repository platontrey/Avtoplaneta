<?php

/**
 * Export a manifest for images referenced by published JoomShopping products.
 *
 * Usage:
 *   php export_joomshopping_photos.php configuration.php img_products_dir optimized|full manifest.csv files.txt
 */

if ($argc !== 6 || !in_array($argv[3], array('optimized', 'full'), true)) {
    fwrite(STDERR, "Usage: php export_joomshopping_photos.php <configuration.php> <img_products_dir> <optimized|full> <manifest.csv> <files.txt>\n");
    exit(2);
}

require_once $argv[1];
$imageDir = rtrim($argv[2], '/');
$variant = $argv[3];
$manifestPath = $argv[4];
$filesPath = $argv[5];

$config = new JConfig();
$prefix = $config->dbprefix;
$db = new mysqli($config->host, $config->user, $config->password, $config->db);
if ($db->connect_errno) {
    fwrite(STDERR, "MySQL connection failed: {$db->connect_error}\n");
    exit(1);
}
$db->set_charset('utf8mb4');

$manifest = fopen($manifestPath, 'wb');
$files = fopen($filesPath, 'wb');
if ($manifest === false || $files === false) {
    fwrite(STDERR, "Unable to create photo export files\n");
    exit(1);
}
fputcsv($manifest, array('product_id', 'photo_url', 'ordering', 'source_file', 'size_bytes'));

$sql = "SELECT i.image_id, p.product_id, i.image_name, i.ordering " .
    "FROM `{$prefix}jshopping_products` p " .
    "LEFT JOIN `{$prefix}jshopping_products_images` i ON i.product_id = p.product_id " .
    "WHERE p.product_publish = 1 " .
    "ORDER BY p.product_id, i.ordering, i.image_id";
$result = $db->query($sql, MYSQLI_USE_RESULT);
if ($result === false) {
    fwrite(STDERR, "Image query failed: {$db->error}\n");
    exit(1);
}

$seenFiles = array();
$manifestRows = 0;
$photoRows = 0;
$missing = 0;
$bytes = 0;
while ($row = $result->fetch_assoc()) {
    $name = basename(trim((string) $row['image_name']));
    if ($name === '' || $name === '.' || $name === '..') {
        fputcsv($manifest, array((int) $row['product_id'], '', 0, '', 0));
        $manifestRows++;
        continue;
    }
    $candidates = $variant === 'full'
        ? array('full_' . $name, $name, 'thumb_' . $name)
        : array($name, 'full_' . $name, 'thumb_' . $name);
    $sourceFile = '';
    foreach ($candidates as $candidate) {
        if (is_file($imageDir . '/' . $candidate)) {
            $sourceFile = $candidate;
            break;
        }
    }
    if ($sourceFile === '') {
        fputcsv($manifest, array(
            (int) $row['product_id'],
            '',
            (int) $row['ordering'],
            '',
            0,
        ));
        $manifestRows++;
        $missing++;
        continue;
    }

    $size = filesize($imageDir . '/' . $sourceFile);
    $size = $size === false ? 0 : (int) $size;
    $photoUrl = '/uploads/joomla/' . rawurlencode($sourceFile);
    fputcsv($manifest, array(
        (int) $row['product_id'],
        $photoUrl,
        (int) $row['ordering'],
        $sourceFile,
        $size,
    ));
    $manifestRows++;
    if (!isset($seenFiles[$sourceFile])) {
        fwrite($files, $sourceFile . "\n");
        $seenFiles[$sourceFile] = true;
        $bytes += $size;
    }
    $photoRows++;
}

$result->free();
fclose($manifest);
fclose($files);
$db->close();
fwrite(STDERR, "manifest_rows={$manifestRows} photo_rows={$photoRows} unique_files=" . count($seenFiles) . " missing={$missing} bytes={$bytes}\n");
