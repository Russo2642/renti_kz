-- Изменение поля booking_number на nullable
-- Это исправляет конфликт с триггером AFTER INSERT
ALTER TABLE bookings ALTER COLUMN booking_number DROP NOT NULL;

-- Обновляем все существующие записи без booking_number
UPDATE bookings 
SET booking_number = 'AD' || LPAD(id::TEXT, 3, '0') 
WHERE booking_number IS NULL; 