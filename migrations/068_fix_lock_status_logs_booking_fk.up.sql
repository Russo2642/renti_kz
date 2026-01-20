-- Исправляем foreign key constraint для booking_id в lock_status_logs
-- Добавляем ON DELETE SET NULL, чтобы при удалении бронирования логи сохранялись

ALTER TABLE lock_status_logs 
DROP CONSTRAINT IF EXISTS lock_status_logs_booking_id_fkey;

ALTER TABLE lock_status_logs 
ADD CONSTRAINT lock_status_logs_booking_id_fkey 
FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE SET NULL;

