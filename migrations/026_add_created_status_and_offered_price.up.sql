-- Добавление нового статуса 'created' в enum статусов бронирования
-- и поля для предложенной пользователем цены

-- Добавляем новый статус 'created' в constraint
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_status_check;
ALTER TABLE bookings ADD CONSTRAINT bookings_status_check 
    CHECK (status IN ('created', 'pending', 'approved', 'rejected', 'active', 'completed', 'canceled'));

-- Добавляем поле для предложенной пользователем цены при бронировании "по договорённости"
ALTER TABLE bookings ADD COLUMN offered_price INTEGER;

-- Обновляем триггерную функцию проверки конфликтов - исключаем статус 'created'
-- так как неподтвержденные бронирования не должны блокировать создание новых
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

-- Добавляем комментарий для документации
COMMENT ON COLUMN bookings.offered_price IS 'Цена, предложенная пользователем при бронировании "по договорённости"';

-- Обновляем описание статусов
COMMENT ON COLUMN bookings.status IS 'Статус бронирования: created - создано, pending - ожидает подтверждения, approved - одобрено, rejected - отклонено, active - активно, completed - завершено, canceled - отменено'; 