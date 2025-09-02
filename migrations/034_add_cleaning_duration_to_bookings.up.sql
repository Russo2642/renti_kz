-- Добавление поля cleaning_duration в таблицу bookings
-- Время на уборку между бронированиями (в минутах)
ALTER TABLE bookings ADD COLUMN cleaning_duration INTEGER NOT NULL DEFAULT 60;

-- Добавление простого индекса для оптимизации запросов
CREATE INDEX idx_bookings_cleaning_duration ON bookings (apartment_id, end_date, cleaning_duration);

-- Комментарий для поля
COMMENT ON COLUMN bookings.cleaning_duration IS 'Время на уборку после окончания бронирования (в минутах). По умолчанию 60 минут.'; 