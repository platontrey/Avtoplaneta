DROP TRIGGER IF EXISTS trigger_notify_auth_outbox ON outbox_events;
DROP FUNCTION IF EXISTS notify_auth_outbox_event();
DROP TABLE IF EXISTS outbox_events;
