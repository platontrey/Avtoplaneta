-- 000004_deduplicate_and_constrain_customers.up.sql

-- 1. Нормализуем телефон и очищаем имена существующих клиентов
UPDATE customers
SET 
    phone = CASE 
        WHEN LENGTH(REGEXP_REPLACE(phone, '[^0-9]', '', 'g')) = 10 THEN 
            '+7' || REGEXP_REPLACE(phone, '[^0-9]', '', 'g')
        WHEN LENGTH(REGEXP_REPLACE(phone, '[^0-9]', '', 'g')) = 11 AND REGEXP_REPLACE(phone, '[^0-9]', '', 'g') ~ '^[78]' THEN 
            '+7' || SUBSTRING(REGEXP_REPLACE(phone, '[^0-9]', '', 'g') FROM 2)
        WHEN LENGTH(REGEXP_REPLACE(phone, '[^0-9]', '', 'g')) >= 10 THEN 
            '+' || REGEXP_REPLACE(phone, '[^0-9]', '', 'g')
        ELSE 
            TRIM(phone)
    END
WHERE phone IS NOT NULL AND TRIM(phone) != '';

-- Очищаем имена от префикса "Получатель" и телефонных номеров в конце строки
UPDATE customers
SET name = TRIM(REGEXP_REPLACE(REGEXP_REPLACE(name, '(?i)^получатель\s*', ''), '\s*\+?[0-9\-\(\)\s]{7,}$', ''))
WHERE name ~* 'получатель' OR name ~ '[0-9]{7,}';

-- Если имя стало пустым после очистки, берем значение телефона
UPDATE customers
SET name = phone
WHERE TRIM(name) = '' AND TRIM(phone) != '';

-- 2. Удаляем мусорные записи без заказов (например, тестовые цифры "1" или служебные фразы)
DELETE FROM customers c
WHERE (c.phone IN ('1', 'Самовывоз', 'Продажа на месте', 'Без контакта') 
       OR c.name IN ('1', 'Самовывоз', 'Продажа на месте', 'Без контакта'))
  AND NOT EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = c.id);

-- 3. Перепривязываем заказы к канонической записи клиента для каждого телефона
WITH ranked_customers AS (
    SELECT 
        c.id,
        c.phone,
        ROW_NUMBER() OVER (
            PARTITION BY TRIM(c.phone) 
            ORDER BY 
                (SELECT COUNT(*) FROM orders o WHERE o.customer_id = c.id) DESC,
                c.id ASC
        ) as rn
    FROM customers c
    WHERE c.phone IS NOT NULL AND TRIM(c.phone) != ''
),
canonical_mapping AS (
    SELECT 
        r.id AS duplicate_id,
        c.id AS canonical_id
    FROM ranked_customers r
    JOIN ranked_customers c ON TRIM(c.phone) = TRIM(r.phone) AND c.rn = 1
    WHERE r.rn > 1
)
UPDATE orders o
SET customer_id = m.canonical_id
FROM canonical_mapping m
WHERE o.customer_id = m.duplicate_id;

-- 4. Привязываем непривязанные заказы (customer_id = 0 или NULL) к существующим клиентам по номеру
WITH canonical AS (
    SELECT phone, id,
           ROW_NUMBER() OVER (PARTITION BY TRIM(phone) ORDER BY id ASC) as rn
    FROM customers
    WHERE phone IS NOT NULL AND TRIM(phone) != ''
)
UPDATE orders o
SET customer_id = c.id
FROM canonical c
WHERE c.rn = 1 
  AND (
      TRIM(o.buyer_number) = c.phone 
      OR ('+7' || REGEXP_REPLACE(o.buyer_number, '[^0-9]', '', 'g')) = c.phone
  )
  AND (o.customer_id = 0 OR o.customer_id IS NULL);

-- 5. Удаляем все дубликаты клиентов с одинаковым телефоном (оставляя только каноническую запись)
WITH ranked_customers AS (
    SELECT 
        id,
        ROW_NUMBER() OVER (
            PARTITION BY TRIM(phone) 
            ORDER BY 
                (SELECT COUNT(*) FROM orders o WHERE o.customer_id = customers.id) DESC,
                id ASC
        ) as rn
    FROM customers
    WHERE phone IS NOT NULL AND TRIM(phone) != ''
)
DELETE FROM customers
WHERE id IN (SELECT id FROM ranked_customers WHERE rn > 1);

-- 6. Удаляем дубликаты клиентов с пустым телефоном, но одинаковым именем
WITH ranked_by_name AS (
    SELECT 
        id,
        ROW_NUMBER() OVER (
            PARTITION BY TRIM(name) 
            ORDER BY 
                (SELECT COUNT(*) FROM orders o WHERE o.customer_id = customers.id) DESC,
                id ASC
        ) as rn
    FROM customers
    WHERE (phone IS NULL OR TRIM(phone) = '') AND TRIM(name) != ''
)
DELETE FROM customers
WHERE id IN (SELECT id FROM ranked_by_name WHERE rn > 1);

-- 7. Создаем уникальный индекс по телефону для предотвращения повторного дублирования
CREATE UNIQUE INDEX IF NOT EXISTS idx_customers_unique_phone
ON customers (TRIM(phone))
WHERE phone IS NOT NULL AND TRIM(phone) != '';
