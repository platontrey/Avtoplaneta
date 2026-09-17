<?php

/**
 * Export published JoomShopping products as a PostgreSQL-friendly CSV file.
 *
 * Usage:
 *   php export_joomshopping.php /path/to/joomla/configuration.php /tmp/joomshopping_parts.csv
 *
 * Product images are intentionally not read or exported.
 */

if ($argc !== 3) {
    fwrite(STDERR, "Usage: php export_joomshopping.php <configuration.php> <output.csv>\n");
    exit(2);
}

$configurationPath = $argv[1];
$outputPath = $argv[2];

if (!is_file($configurationPath)) {
    fwrite(STDERR, "Joomla configuration file not found: {$configurationPath}\n");
    exit(2);
}

require_once $configurationPath;
$config = new JConfig();
$prefix = $config->dbprefix;

$db = new mysqli($config->host, $config->user, $config->password, $config->db);
if ($db->connect_errno) {
    fwrite(STDERR, "MySQL connection failed: {$db->connect_error}\n");
    exit(1);
}
$db->set_charset('utf8mb4');

$values = [];
$valuesResult = $db->query(
    "SELECT id, field_id, `name_ru-RU` AS value_name " .
    "FROM `{$prefix}jshopping_products_extra_field_values`"
);
while ($row = $valuesResult->fetch_assoc()) {
    $values[(int) $row['field_id']][(string) $row['id']] = trim((string) $row['value_name']);
}
$valuesResult->free();

$categories = [];
$categoriesResult = $db->query(
    "SELECT category_id, category_parent_id, `name_ru-RU` AS category_name " .
    "FROM `{$prefix}jshopping_categories`"
);
while ($row = $categoriesResult->fetch_assoc()) {
    $categories[(string) $row['category_id']] = array(
        'parent_id' => (string) $row['category_parent_id'],
        'name' => trim((string) $row['category_name']),
    );
}
$categoriesResult->free();

$partCategories = array(
    'Выхлопная система',
    'Двигатель',
    'Диски и шины',
    'Кузов внутри',
    'Кузов снаружи',
    'Оптика',
    'Пневмосистема',
    'Подвеска ДВС/КПП',
    'Подвеска задних колес',
    'Подвеска передних колес',
    'Рулевое управление',
    'Система кондиционирования',
    'Система охлаждения и отопления',
    'Сопутствующие товары',
    'Стекла',
    'Тормозная система',
    'Трансмиссия',
    'Электрооснащение',
);

function cleanScalar($value)
{
    $value = (string) (isset($value) ? $value : '');
    $value = str_replace("\0", '', $value);
    // Legacy values sometimes use backslashes in addresses and before quotes.
    // PHP 5.4's CSV parser treats those as enclosure escapes, so normalize them.
    $value = str_replace('\\', '/', $value);
    return trim($value);
}

function resolveExtra($values, $fieldId, $rawValue)
{
    $raw = cleanScalar($rawValue);
    if ($raw === '' || $raw === '0') {
        return '';
    }
    return isset($values[$fieldId][$raw]) ? $values[$fieldId][$raw] : $raw;
}

function isMalformedLegacyName($value)
{
    // Some Joomla records contain a comma-separated vehicle specification
    // instead of a part name, e.g. "HYUNDAI, SANTA FE, ..., 04.2007 -
    // 04.2013, ...". The first field varies, so do not rely on it being "0".
    $name = cleanScalar($value);
    return substr_count($name, ',') >= 5
        && preg_match('/\b[0-9]{2}\.[0-9]{4}\s*-\s*[0-9]{2}\.[0-9]{4}\b/u', $name) === 1;
}

function normalizeVehicleText($value)
{
    $value = cleanScalar($value);
    if ($value === '' || !function_exists('mb_strtolower')) {
        return $value;
    }

    // Legacy Joomla stores many makes/models in all caps. Convert only those
    // values, preserving mixed-case model codes and known automotive acronyms.
    if (preg_match('/^[A-ZА-ЯЁ0-9 ._\/-]+$/u', $value)) {
        $value = mb_convert_case(mb_strtolower($value, 'UTF-8'), MB_CASE_TITLE, 'UTF-8');
        foreach (['Bmw' => 'BMW', 'Gmc' => 'GMC', 'Daf' => 'DAF', 'Man' => 'MAN', 'Ud' => 'UD'] as $from => $to) {
            $value = preg_replace('/\b' . preg_quote($from, '/') . '\b/u', $to, $value);
        }
    }
    return $value;
}

function resolveAddress($values, $rawValue)
{
    $raw = cleanScalar($rawValue);
    $historicalAliases = array(
        // Deleted Joomla options recovered from the 2022 SQL backup.
        '1239' => 'г. Томск ул. 79-й Гвардейской Дивизии, 23/1',
        '1336' => 'г. Томск ул. 79-й Гвардейской Дивизии, 23/1',
        // Older deleted options inferred from the actual shelf prefixes.
        '878' => 'г. Томск ул. 79-й Гвардейской Дивизии, 23/1',
        '880' => 'г. Томск, пос. Кузовлево, ул. Заречная, 36',
    );
    if (isset($historicalAliases[$raw])) {
        return $historicalAliases[$raw];
    }
    return cleanScalar(resolveExtra($values, 44, $raw));
}

function cleanDescription($parts)
{
    $unique = [];
    foreach ($parts as $part) {
        $text = html_entity_decode(strip_tags(cleanScalar($part)), ENT_QUOTES | ENT_HTML5, 'UTF-8');
        $normalized = preg_replace('/\s+/u', ' ', $text);
        $text = $normalized !== null ? $normalized : $text;
        $text = trim($text);
        if ($text !== '' && !in_array($text, $unique, true)) {
            $unique[] = $text;
        }
    }
    $description = implode(' | ', $unique);
    return function_exists('mb_substr') ? mb_substr($description, 0, 1000, 'UTF-8') : substr($description, 0, 1000);
}

function validDateTime($value, $fallback)
{
    $value = cleanScalar($value);
    if ($value === '' || strpos($value, '0000-00-00') === 0) {
        return $fallback;
    }
    return $value;
}

function resolveCategory($values, $rawValue, $legacyCategoryId, $categories, $partCategories)
{
    $category = resolveExtra($values, 52, $rawValue);
    if ($category !== '') {
        return $category;
    }

    $categoryId = cleanScalar($legacyCategoryId);
    $visited = array();
    while ($categoryId !== '' && $categoryId !== '0' && isset($categories[$categoryId]) && count($visited) < 20) {
        if (isset($visited[$categoryId])) {
            break;
        }
        $visited[$categoryId] = true;
        $legacyName = cleanScalar($categories[$categoryId]['name']);
        foreach ($partCategories as $partCategory) {
            if (stripos($legacyName, $partCategory) === 0) {
                return $partCategory;
            }
        }
        $categoryId = cleanScalar($categories[$categoryId]['parent_id']);
    }

    return '';
}

$output = fopen($outputPath, 'wb');
if ($output === false) {
    fwrite(STDERR, "Cannot open output file: {$outputPath}\n");
    exit(1);
}

$columns = [
    'id', 'name', 'quantity', 'description', 'category', 'price', 'salesman', 'location', 'address', 'status',
    'brand', 'model', 'photos', 'seller_id', 'vin', 'body_brand', 'engine_brand', 'car_release_date',
    'front_rear', 'left_right', 'top_bottom', 'number', 'manufacturer', 'manufacturer_code', 'oem_code',
    'color', 'condition', 'supplier_code', 'defect', 'transmission', 'transmission_model', 'drive',
    'wear_percentage', 'season', 'diameter', 'width', 'profile', 'tire_quantity', 'drilling', 'offset',
    'center_hole_diameter', 'tire_model', 'created_at', 'updated_at',
];
fputcsv($output, $columns, ',', '"');

$selectColumns = [
    'product_id', 'product_ean', 'product_quantity', 'product_date_added', 'date_modify',
    'product_publish', 'product_price', 'name_ru-RU', 'short_description_ru-RU', 'description_ru-RU',
    'extra_field_29', 'extra_field_30', 'extra_field_31', 'extra_field_32', 'extra_field_33',
    'extra_field_34', 'extra_field_35', 'extra_field_36', 'extra_field_37', 'extra_field_38',
    'extra_field_39', 'extra_field_40', 'extra_field_41', 'extra_field_42', 'extra_field_43',
    'extra_field_44', 'extra_field_45', 'extra_field_48', 'extra_field_49', 'extra_field_50',
    'extra_field_52', 'extra_field_55', 'extra_field_56', 'extra_field_57', 'extra_field_58',
    'extra_field_61', 'extra_field_62', 'extra_field_65', 'extra_field_66', 'extra_field_67',
    'extra_field_68', 'extra_field_69', 'extra_field_70', 'extra_field_71', 'extra_field_72',
    'extra_field_73',
];
$quotedColumns = array_map(function ($column) {
    return "p.`{$column}`";
}, $selectColumns);
$sql = "SELECT " . implode(', ', $quotedColumns) . ", pc.legacy_category_id" .
    " FROM `{$prefix}jshopping_products` p" .
    " LEFT JOIN (" .
        "SELECT product_id, MIN(category_id) AS legacy_category_id " .
        "FROM `{$prefix}jshopping_products_to_categories` GROUP BY product_id" .
    ") pc ON pc.product_id = p.product_id" .
    " WHERE p.product_publish = 1 ORDER BY p.product_id";
$products = $db->query($sql, MYSQLI_USE_RESULT);
if ($products === false) {
    fwrite(STDERR, "Product query failed: {$db->error}\n");
    exit(1);
}

$count = 0;
$skippedMalformedNames = 0;
$now = gmdate('Y-m-d H:i:s');
while ($row = $products->fetch_assoc()) {
    $quantity = max(0, (int) floor((float) $row['product_quantity']));
    $name = resolveExtra($values, 29, $row['extra_field_29']);
    if ($name === '') {
        $name = cleanScalar($row['name_ru-RU']);
    }
    if (isMalformedLegacyName($name)) {
        fwrite(STDERR, "Skipping product {$row['product_id']}: malformed legacy name\n");
        ++$skippedMalformedNames;
        continue;
    }

    $address = resolveAddress($values, $row['extra_field_44']);
    $shelf = resolveExtra($values, 62, $row['extra_field_62']);
    if ($shelf === '') {
        $shelf = resolveExtra($values, 61, $row['extra_field_61']);
    }

    $number = resolveExtra($values, 38, $row['extra_field_38']);
    if ($number === '') {
        $number = cleanScalar($row['product_ean']);
    }

    $createdAt = validDateTime($row['product_date_added'], $now);
    $updatedAt = validDateTime($row['date_modify'], $createdAt);
    $category = resolveCategory(
        $values,
        $row['extra_field_52'],
        $row['legacy_category_id'],
        $categories,
        $partCategories
    );
    if ($category === '') {
        $normalizedName = function_exists('mb_strtolower') ? mb_strtolower($name, 'UTF-8') : strtolower($name);
        if (strpos($normalizedName, 'шин') !== false) {
            $category = 'Диски и шины';
        } elseif (strpos($normalizedName, 'webasto') !== false) {
            $category = 'Система охлаждения и отопления';
        }
    }

    $exportRow = [
        (string) $row['product_id'],
        $name,
        (string) $quantity,
        cleanDescription([
            $row['extra_field_45'],
            $row['extra_field_57'],
            $row['short_description_ru-RU'],
            $row['description_ru-RU'],
        ]),
        $category,
        number_format((float) $row['product_price'], 6, '.', ''),
        resolveExtra($values, 49, $row['extra_field_49']),
        $shelf,
        $address,
        $quantity > 0 ? 'true' : 'false',
        normalizeVehicleText(resolveExtra($values, 30, $row['extra_field_30'])),
        normalizeVehicleText(resolveExtra($values, 31, $row['extra_field_31'])),
        '[]',
        '0',
        '',
        resolveExtra($values, 32, $row['extra_field_32']),
        resolveExtra($values, 33, $row['extra_field_33']),
        resolveExtra($values, 34, $row['extra_field_34']),
        resolveExtra($values, 35, $row['extra_field_35']),
        resolveExtra($values, 36, $row['extra_field_36']),
        resolveExtra($values, 37, $row['extra_field_37']),
        $number,
        resolveExtra($values, 39, $row['extra_field_39']),
        resolveExtra($values, 40, $row['extra_field_40']),
        resolveExtra($values, 41, $row['extra_field_41']),
        resolveExtra($values, 42, $row['extra_field_42']),
        resolveExtra($values, 43, $row['extra_field_43']),
        resolveExtra($values, 48, $row['extra_field_48']),
        resolveExtra($values, 50, $row['extra_field_50']),
        resolveExtra($values, 55, $row['extra_field_55']),
        '',
        resolveExtra($values, 56, $row['extra_field_56']),
        resolveExtra($values, 58, $row['extra_field_58']),
        resolveExtra($values, 65, $row['extra_field_65']),
        resolveExtra($values, 66, $row['extra_field_66']),
        resolveExtra($values, 67, $row['extra_field_67']),
        resolveExtra($values, 68, $row['extra_field_68']),
        resolveExtra($values, 69, $row['extra_field_69']),
        resolveExtra($values, 70, $row['extra_field_70']),
        resolveExtra($values, 71, $row['extra_field_71']),
        resolveExtra($values, 72, $row['extra_field_72']),
        resolveExtra($values, 73, $row['extra_field_73']),
        $createdAt,
        $updatedAt,
    ];

    fputcsv($output, $exportRow, ',', '"');
    ++$count;
}

$products->free();
fclose($output);
$db->close();

fwrite(STDOUT, "Exported {$count} published products to {$outputPath} (photos excluded; skipped malformed names={$skippedMalformedNames}).\n");
