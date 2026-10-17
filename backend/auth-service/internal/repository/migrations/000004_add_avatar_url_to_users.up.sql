-- Добавление колонки avatar_url для хранения пути к аватару пользователя
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT NOT NULL DEFAULT '';
