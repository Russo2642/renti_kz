-- Откат добавления payment_id в booking_extensions

-- Удаляем индекс
DROP INDEX IF EXISTS idx_booking_extensions_payment_id;

-- Возвращаем старый constraint для статусов
ALTER TABLE booking_extensions DROP CONSTRAINT IF EXISTS booking_extensions_status_check;
ALTER TABLE booking_extensions ADD CONSTRAINT booking_extensions_status_check 
    CHECK (status IN ('pending', 'approved', 'rejected'));

-- Удаляем поле payment_id
ALTER TABLE booking_extensions DROP COLUMN IF EXISTS payment_id; 