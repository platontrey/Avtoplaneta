-- 1. Добавление колонки username
ALTER TABLE users ADD COLUMN IF NOT EXISTS username TEXT NOT NULL DEFAULT '';

-- 2. Заполнение username для существующих пользователей на основе email
UPDATE users
SET username = CASE 
    WHEN email != '' AND email IS NOT NULL THEN LOWER(SPLIT_PART(email, '@', 1))
    ELSE 'user_' || id::text
END
WHERE username = '';

-- Разрешение коллизий если префиксы email совпадали
UPDATE users u
SET username = u.username || '_' || u.id::text
WHERE u.id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER(PARTITION BY LOWER(TRIM(username)) ORDER BY id) as rn
        FROM users
    ) t WHERE t.rn > 1
);

-- 3. Уникальный индекс по логину (без учета регистра)
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_lower ON users (LOWER(TRIM(username))) WHERE username != '';

-- 4. Снятие жесткого NOT NULL с email для поддержки пользователей только с логином
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
ALTER TABLE users ALTER COLUMN email SET DEFAULT '';

-- 5. Замена строгого users_email_key на частичный уникальный индекс, чтобы пустые email не конфликтовали
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
DROP INDEX IF EXISTS idx_users_email_lower;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(TRIM(email))) WHERE email IS NOT NULL AND email != '';
