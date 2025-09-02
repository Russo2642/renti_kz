-- Откат добавления статуса 'created' и поля offered_price

-- Удаляем поле offered_price
ALTER TABLE bookings DROP COLUMN IF EXISTS offered_price;

-- Возвращаем старый constraint статусов (без 'created')
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_status_check;
ALTER TABLE bookings ADD CONSTRAINT bookings_status_check 
    CHECK (status IN ('pending', 'approved', 'rejected', 'active', 'completed', 'canceled'));

-- Возвращаем старое описание статусов
COMMENT ON COLUMN bookings.status IS 'Статус бронирования: pending - ожидает подтверждения, approved - одобрено, rejected - отклонено, active - активно, completed - завершено, canceled - отменено'; 