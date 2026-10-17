-- Заполняем пустые номера сделок идентификатором заказа
UPDATE orders 
SET order_number = id::text 
WHERE order_number = '' OR order_number IS NULL;
