<?php

/**
 * Compare the legacy JoomShopping shelf fields:
 *   extra_field_61 ("Полка1", select/value ID)
 *   extra_field_62 ("Полка", free text)
 *
 * Compatible with the PHP 5.4 CLI on the legacy server.
 */

if ($argc !== 2) {
    fwrite(STDERR, "Usage: php audit_joomshopping_shelves.php <configuration.php>\n");
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

$shelfValues = array();
$values = $db->query(
    "SELECT id, `name_ru-RU` AS shelf FROM `{$prefix}jshopping_products_extra_field_values` WHERE field_id = 61"
);
while ($row = $values->fetch_assoc()) {
    $shelfValues[(string) $row['id']] = trim((string) $row['shelf']);
}
$values->free();

function shelfText($value)
{
    $value = trim(str_replace('\\', '/', (string) $value));
    return $value === '0' ? '' : $value;
}

function normalizedShelf($value)
{
    $value = shelfText($value);
    $normalized = preg_replace('/\s+/u', '', $value);
    if ($normalized !== null) {
        $value = $normalized;
    }
    return function_exists('mb_strtoupper') ? mb_strtoupper($value, 'UTF-8') : strtoupper($value);
}

function appendSample(&$samples, $row, $shelf1, $shelf)
{
    if (count($samples) >= 12) {
        return;
    }
    $samples[] = array(
        'product_id' => (int) $row['product_id'],
        'date' => (string) $row['product_date_added'],
        'polka1_raw' => (string) $row['extra_field_61'],
        'polka1_resolved' => $shelf1,
        'polka' => $shelf,
    );
}

$stats = array(
    'total' => 0,
    'neither' => 0,
    'only_polka1' => 0,
    'only_polka' => 0,
    'both' => 0,
    'both_equal' => 0,
    'both_different' => 0,
    'unresolved_polka1_ids' => 0,
);
$byYear = array();
$samples = array(
    'only_polka1' => array(),
    'only_polka' => array(),
    'both_different' => array(),
);

$products = $db->query(
    "SELECT product_id, product_date_added, extra_field_61, extra_field_62 " .
    "FROM `{$prefix}jshopping_products` WHERE product_publish = 1 ORDER BY product_id",
    MYSQLI_USE_RESULT
);
while ($row = $products->fetch_assoc()) {
    ++$stats['total'];
    $rawShelf1 = shelfText($row['extra_field_61']);
    $shelf1 = $rawShelf1 === '' ? '' : (isset($shelfValues[$rawShelf1]) ? $shelfValues[$rawShelf1] : $rawShelf1);
    $shelf = shelfText($row['extra_field_62']);
    if ($rawShelf1 !== '' && !isset($shelfValues[$rawShelf1]) && ctype_digit($rawShelf1)) {
        ++$stats['unresolved_polka1_ids'];
    }

    $year = substr((string) $row['product_date_added'], 0, 4);
    if (!isset($byYear[$year])) {
        $byYear[$year] = array('total' => 0, 'only_polka1' => 0, 'only_polka' => 0, 'both_equal' => 0, 'both_different' => 0, 'neither' => 0);
    }
    ++$byYear[$year]['total'];

    if ($shelf1 === '' && $shelf === '') {
        ++$stats['neither'];
        ++$byYear[$year]['neither'];
    } elseif ($shelf1 !== '' && $shelf === '') {
        ++$stats['only_polka1'];
        ++$byYear[$year]['only_polka1'];
        appendSample($samples['only_polka1'], $row, $shelf1, $shelf);
    } elseif ($shelf1 === '' && $shelf !== '') {
        ++$stats['only_polka'];
        ++$byYear[$year]['only_polka'];
        appendSample($samples['only_polka'], $row, $shelf1, $shelf);
    } else {
        ++$stats['both'];
        if (normalizedShelf($shelf1) === normalizedShelf($shelf)) {
            ++$stats['both_equal'];
            ++$byYear[$year]['both_equal'];
        } else {
            ++$stats['both_different'];
            ++$byYear[$year]['both_different'];
            appendSample($samples['both_different'], $row, $shelf1, $shelf);
        }
    }
}
$products->free();
ksort($byYear);

echo json_encode(
    array('stats' => $stats, 'by_year' => $byYear, 'samples' => $samples),
    JSON_UNESCAPED_UNICODE | JSON_PRETTY_PRINT
) . "\n";

