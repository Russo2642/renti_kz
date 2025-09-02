-- Добавление поля payment_id в таблицу booking_extensions
-- для связи продлений с платежами

-- Добавляем поле payment_id от таблицы payments
ALTER TABLE booking_extensions ADD COLUMN payment_id BIGINT REFERENCES payments(id);

-- Обновляем constraint для статусов продлений, добавляем awaiting_payment
ALTER TABLE booking_extensions DROP CONSTRAINT IF EXISTS booking_extensions_status_check;
ALTER TABLE booking_extensions ADD CONSTRAINT booking_extensions_status_check 
    CHECK (status IN ('awaiting_payment', 'pending', 'approved', 'rejected'));

-- Создаем индекс для быстрого поиска продлений по payment_id
CREATE INDEX idx_booking_extensions_payment_id ON booking_extensions(payment_id) WHERE payment_id IS NOT NULL; 