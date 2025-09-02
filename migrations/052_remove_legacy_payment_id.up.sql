-- Удаление legacy payment_id колонки из bookings
-- 052_remove_legacy_payment_id.up.sql

-- Убеждаемся, что все записи имеют payment_db_id
UPDATE bookings 
SET payment_db_id = (
    SELECT p.id 
    FROM payments p 
    WHERE p.payment_id = bookings.payment_id
    AND p.status = 'success'
    ORDER BY p.created_at DESC 
    LIMIT 1
)
WHERE payment_db_id IS NULL 
AND payment_id IS NOT NULL;

-- Удаляем legacy колонку
ALTER TABLE bookings DROP COLUMN payment_id; 