# Part catalog

`catalog.json` is the single source of truth for ordinary-part forms and defect-report templates. The parts service embeds it and exposes it through `GET /api/part-catalog`; the web and Flutter clients consume the same response.

To change the catalog:

1. Update `version` so clients revalidate their cached copy.
2. Edit `attributes` for field metadata.
3. Edit `part_form_categories` to control which fields appear for each ordinary-part category.
4. Edit `report_bindings` to control which vehicle values are copied to which defect-report part categories.
5. Edit `parts` to add, remove, or change defect-report templates.
6. Run `go test ./...` from `backend/parts-service`.

The service validates duplicate IDs, unknown fields, and dependency categories during startup. Keep category names in `report_bindings` identical to categories used by the part templates.

`TestDefectReportHTTPThroughRedisConsumer` covers the complete in-process defect-report route: HTTP request, catalog expansion, Redis Stream delivery, consumer-group processing, creation of all 1,442 parts, specification bindings, and stream acknowledgement. It uses an in-memory Redis server and a recording inventory service, so it never writes test parts to PostgreSQL.
