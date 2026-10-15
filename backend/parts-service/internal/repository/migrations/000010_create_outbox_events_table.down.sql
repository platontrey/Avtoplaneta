DROP TRIGGER IF EXISTS trigger_notify_outbox_event ON outbox_events;
DROP FUNCTION IF EXISTS notify_outbox_event();
DROP TABLE IF EXISTS outbox_events;
