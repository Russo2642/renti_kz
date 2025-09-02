-- Добавление нового статуса 'awaiting_payment' в enum статусов бронирования
-- и поля для payment_id

-- Добавляем новый статус 'awaiting_payment' в constraint
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_status_check;
ALTER TABLE bookings ADD CONSTRAINT bookings_status_check 
    CHECK (status IN ('created', 'awaiting_payment', 'pending', 'approved', 'rejected', 'active', 'completed', 'canceled'));

-- Добавляем поле для payment_id от FreedomPay
ALTER TABLE bookings ADD COLUMN payment_id VARCHAR(255);

-- Создаем индекс для быстрого поиска броней по payment_id
CREATE INDEX idx_bookings_payment_id ON bookings(payment_id) WHERE payment_id IS NOT NULL;

-- Обновляем триггерную функцию проверки конфликтов - добавляем исключение для 'awaiting_payment'
-- так как брони ожидающие оплаты не должны блокировать создание новых
CREATE OR REPLACE FUNCTION check_booking_conflicts() RETURNS trigger AS $$
BEGIN
    -- Проверяем конфликты только для активных статусов бронирований
    -- Исключаем 'created' и 'awaiting_payment' так как это неоплаченные заявки
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