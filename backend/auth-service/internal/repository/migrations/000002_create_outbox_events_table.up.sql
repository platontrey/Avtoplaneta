CREATE TABLE IF NOT EXISTS outbox_events (
    id BIGSERIAL PRIMARY KEY,
    aggregate_type VARCHAR(64) NOT NULL DEFAULT 'user',
    aggregate_id BIGINT NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    retry_count INT NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_auth_outbox_pending ON outbox_events(status, created_at) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_auth_outbox_aggregate ON outbox_events(aggregate_type, aggregate_id);

-- Триггер для мгновенного пробуждения воркера через LISTEN / NOTIFY
CREATE OR REPLACE FUNCTION notify_auth_outbox_event()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('auth_outbox_channel', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_notify_auth_outbox ON outbox_events;
CREATE TRIGGER trigger_notify_auth_outbox
AFTER INSERT ON outbox_events
FOR EACH ROW
EXECUTE FUNCTION notify_auth_outbox_event();
