-- Откат: возвращаем constraint без ON DELETE SET NULL

ALTER TABLE lock_status_logs 
DROP CONSTRAINT IF EXISTS lock_status_logs_booking_id_fkey;

ALTER TABLE lock_status_logs 
ADD CONSTRAINT lock_status_logs_booking_id_fkey 
FOREIGN KEY (booking_id) REFERENCES bookings(id);

