-- Переименование payment_db_id в payment_id
-- 053_rename_payment_db_id_to_payment_id.up.sql

-- Переименовываем колонку
ALTER TABLE bookings RENAME COLUMN payment_db_id TO payment_id;

-- Переименовываем индекс
DROP INDEX IF EXISTS idx_bookings_payment_db_id;
CREATE INDEX idx_bookings_payment_id ON bookings(payment_id);

-- Переименовываем FK constraint
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS fk_bookings_payment;
ALTER TABLE bookings ADD CONSTRAINT fk_bookings_payment 
FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL; 