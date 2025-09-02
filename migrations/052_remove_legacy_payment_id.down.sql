-- Восстановление legacy payment_id колонки (для отката)
-- 052_remove_legacy_payment_id.down.sql

-- Добавляем обратно колонку
ALTER TABLE bookings ADD COLUMN payment_id VARCHAR(255);

-- Заполняем данными из payments таблицы
UPDATE bookings 
SET payment_id = (
    SELECT p.payment_id 
    FROM payments p 
    WHERE p.id = bookings.payment_db_id
)
WHERE payment_db_id IS NOT NULL; 