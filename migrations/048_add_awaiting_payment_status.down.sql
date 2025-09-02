-- Откат добавления статуса 'awaiting_payment' и поля payment_id

-- Удаляем индекс
DROP INDEX IF EXISTS idx_bookings_payment_id;

-- Удаляем поле payment_id
ALTER TABLE bookings DROP COLUMN IF EXISTS payment_id;

-- Возвращаем старый constraint без 'awaiting_payment'
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_status_check;
ALTER TABLE bookings ADD CONSTRAINT bookings_status_check 
    CHECK (status IN ('created', 'pending', 'approved', 'rejected', 'active', 'completed', 'canceled'));

-- Возвращаем старую триггерную функцию проверки конфликтов
CREATE OR REPLACE FUNCTION check_booking_conflicts() RETURNS trigger AS $$
BEGIN
    -- Проверяем конфликты только для активных статусов бронирований
    -- Исключаем 'created' так как это неподтвержденные заявки
    IF NEW.status IN ('pending', 'approved', 'active') THEN
        -- Проверяем наличие пересекающихся бронирований
        IF EXISTS (
            SELECT 1 FROM bookings 
            WHERE apartment_id = NEW.apartment_id 
            AND id != COALESCE(NEW.id, 0)
            AND status IN ('pending', 'approved', 'active')
            AND daterange_overlaps(start_date, end_date, NEW.start_date, NEW.end_date)
        ) THEN
            RAISE EXCEPTION 'Apartment is not available for the selected period. Conflicting booking exists.';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql; 