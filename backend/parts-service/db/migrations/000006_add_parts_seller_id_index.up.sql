-- Имя продавца в parts.salesman — производная копия от users.name, нужная для
-- поиска: фильтр «по продавцу» уходит в Elasticsearch, где лежит копия из строки.
-- После переименования копию надо чинить одним UPDATE по seller_id, а для этого
-- нужен индекс: без него правка имени означает полное сканирование parts.
CREATE INDEX IF NOT EXISTS idx_parts_seller_id ON parts (seller_id) WHERE deleted_at IS NULL;
