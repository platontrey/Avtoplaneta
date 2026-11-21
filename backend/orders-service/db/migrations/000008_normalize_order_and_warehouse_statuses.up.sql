-- Миграция: приведение legacy-статусов к единому стандарту
UPDATE orders
SET status = 'completed'
WHERE status IN ('green', 'Выдан / Завершён', 'Выдан клиенту', 'Завершён')
   OR auto_deleted = TRUE;

UPDATE orders
SET status = 'ready'
WHERE status IN ('Готов к выдаче', 'Проверен');

UPDATE orders
SET status = 'cancelled'
WHERE status IN ('Отменен', 'Возврат');

UPDATE orders
SET status = 'processing'
WHERE status IN ('red', 'brown', 'yellow', 'Принят в обработку', 'Ожидает предоплаты', 'На фотофиксации', 'Перемещение между складами', 'Упаковывается');

UPDATE orders
SET warehouse_status = 'pending'
WHERE warehouse_status IS NULL OR warehouse_status = '';

UPDATE orders
SET warehouse_status = 'packed'
WHERE warehouse_status = 'ready';

UPDATE orders
SET warehouse_status = 'picking'
WHERE warehouse_status = 'transfer';
