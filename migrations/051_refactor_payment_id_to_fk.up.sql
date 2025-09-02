-- Рефакторинг payment_id в bookings с VARCHAR на FK
-- 051_refactor_payment_id_to_fk.up.sql

-- 1. Добавляем новое поле payment_db_id как FK
ALTER TABLE bookings ADD COLUMN payment_db_id INTEGER;

-- 2. Заполняем payment_db_id на основе существующих данных
UPDATE bookings 
SET payment_db_id = (
    SELECT p.id 
    FROM payments p 
    WHERE p.payment_id = bookings.payment_id
    AND p.status = 'success'
    ORDER BY p.created_at DESC 
    LIMIT 1
)
WHERE payment_id IS NOT NULL;

-- 3. Создаем FK constraint
ALTER TABLE bookings 
ADD CONSTRAINT fk_bookings_payment 
FOREIGN KEY (payment_db_id) REFERENCES payments(id) ON DELETE SET NULL;

-- 4. Создаем индекс для производительности
CREATE INDEX idx_bookings_payment_db_id ON bookings(payment_db_id);

-- 5. После тестирования можно будет удалить старое поле:
-- ALTER TABLE bookings DROP COLUMN payment_id;
-- (пока оставляем для обратной совместимости) 