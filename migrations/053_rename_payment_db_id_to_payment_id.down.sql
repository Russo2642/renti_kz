-- Откат переименования payment_id обратно в payment_db_id
-- 053_rename_payment_db_id_to_payment_id.down.sql

-- Откатываем FK constraint
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS fk_bookings_payment;
ALTER TABLE bookings ADD CONSTRAINT fk_bookings_payment 
FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL;

-- Откатываем индекс
DROP INDEX IF EXISTS idx_bookings_payment_id;
CREATE INDEX idx_bookings_payment_db_id ON bookings(payment_id);

-- Откатываем переименование колонки
ALTER TABLE bookings RENAME COLUMN payment_id TO payment_db_id; 