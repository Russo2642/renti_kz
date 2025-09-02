-- Откат изменения поля booking_number обратно на NOT NULL
-- Сначала убеждаемся, что все записи имеют booking_number
UPDATE bookings 
SET booking_number = 'AD' || LPAD(id::TEXT, 3, '0') 
WHERE booking_number IS NULL;

-- Устанавливаем обратно NOT NULL constraint
ALTER TABLE bookings ALTER COLUMN booking_number SET NOT NULL; 