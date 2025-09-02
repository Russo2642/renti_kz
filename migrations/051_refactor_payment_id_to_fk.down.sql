-- Откат рефакторинга payment_id 
-- 051_refactor_payment_id_to_fk.down.sql

-- 1. Удаляем индекс
DROP INDEX IF EXISTS idx_bookings_payment_db_id;

-- 2. Удаляем FK constraint
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS fk_bookings_payment;

-- 3. Удаляем новое поле
ALTER TABLE bookings DROP COLUMN IF EXISTS payment_db_id; 