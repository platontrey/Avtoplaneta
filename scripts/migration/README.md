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

## GitHub Actions

The `Migrate Joomla catalog` workflow runs manually on the target server's
self-hosted runner. It always exports and validates both parts and the matching
photo manifest. Available modes are:

- `dry-run`: export and validate without changing the target;
- `data-only`: back up and upsert parts while preserving existing photos;
- `data-and-photos`: upsert parts, rsync referenced images into the persistent
  `parts-uploads` Docker volume, import photo URLs, and rebuild Elasticsearch.

The photo variant can be `optimized` (the normal JoomShopping image, recommended)
or `full` (`full_` original with a fallback to the normal image). Existing files
are updated idempotently and unrelated uploads are never deleted. Rsync runs in
an ephemeral Alpine toolbox container with the named upload volume mounted at
`/dest`; host permissions under `/var/lib/docker/volumes` are not weakened.

Configure these repository settings before running the workflow:

- secret `JOOMLA_SSH_PASSWORD` (required);
- secret `JOOMLA_SSH_KNOWN_HOSTS` (recommended; output of `ssh-keyscan -H`);
- variable `JOOMLA_SSH_HOST` (defaults to the legacy host in the workflow);
- variable `JOOMLA_SSH_USER` (defaults to `root`).

The `data-and-photos` mode additionally requires the dispatch confirmation
`MIGRATE_WITH_PHOTOS`. Every mutating run creates a PostgreSQL custom-format
backup under `/home/avtoplaneta/backups` before importing.
