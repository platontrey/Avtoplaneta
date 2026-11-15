CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL DEFAULT '',
    initials TEXT NOT NULL DEFAULT '',
    inn TEXT NOT NULL DEFAULT '',
    provider TEXT NOT NULL DEFAULT 'local',
    role TEXT NOT NULL DEFAULT 'operator',
    password TEXT NOT NULL DEFAULT ''
) WITH (fillfactor = 90);

CREATE TABLE IF NOT EXISTS user_activity_logs (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL,
    resource_id   BIGINT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    action        TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    user_name     TEXT NOT NULL,
    user_email    TEXT NOT NULL,
    details       TEXT,
    ip_address    TEXT,
    user_agent    TEXT
);

CREATE INDEX IF NOT EXISTS idx_user_activity_logs_user_id ON user_activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_created_at_brin ON user_activity_logs USING BRIN (created_at);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_created_at ON user_activity_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_action ON user_activity_logs(action);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_resource_type ON user_activity_logs(resource_type);
