DROP INDEX IF EXISTS idx_users_username_lower;
DROP INDEX IF EXISTS idx_users_email_lower;

UPDATE users SET email = username || '@local' WHERE email IS NULL OR email = '';
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE users DROP COLUMN IF EXISTS username;
