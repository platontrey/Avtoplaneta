# JoomShopping catalog migration

This migration moves published products from the legacy Joomla/JoomShopping MySQL database into the project's PostgreSQL `parts` table.

- Source product IDs are preserved.
- JoomShopping extra-field IDs are resolved to their Russian display values.
- Only published products are exported.
- Product photos and image references are deliberately excluded (`photos = []`).
- Re-running the import updates rows with the same source ID.
- `extra_field_62` is the authoritative shelf; resolved `extra_field_61` is used
  only when field 62 is empty.
- Deleted legacy address options are normalized to the three current warehouse
  addresses. IDs 1239 and 1336 were recovered from an older SQL backup; IDs 878
  and 880 are mapped by the matching shelf prefixes in the affected products.

The exporter reads database credentials directly from Joomla's `configuration.php`; credentials are not copied into the repository or command output.

```bash
php export_joomshopping.php /path/to/joomla/configuration.php /tmp/joomshopping_parts.csv
```

Copy the CSV and `import_joomshopping.sql` into the PostgreSQL container as `/tmp/joomshopping_parts.csv` and `/tmp/import_joomshopping.sql`, then run:

```bash
psql -v ON_ERROR_STOP=1 -f /tmp/import_joomshopping.sql
```

Restart the parts service after importing so the Elasticsearch index is rebuilt.
