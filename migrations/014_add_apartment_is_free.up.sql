-- Добавляем поле is_free для отслеживания доступности квартиры
ALTER TABLE apartments ADD COLUMN is_free BOOLEAN NOT NULL DEFAULT true;

-- Создаем индекс для быстрого поиска свободных квартир
CREATE INDEX idx_apartments_is_free ON apartments(is_free);

-- Создаем составной индекс для фильтрации по статусу и доступности
CREATE INDEX idx_apartments_status_is_free ON apartments(status, is_free);

-- Обновляем существующие квартиры: если есть активные бронирования, то квартира занята
UPDATE apartments 
SET is_free = false 
WHERE id IN (
    SELECT DISTINCT apartment_id 
    FROM bookings 
    WHERE status = 'active'
);

-- Комментарий к полю
COMMENT ON COLUMN apartments.is_free IS 'Флаг доступности квартиры для бронирования (true - свободна, false - занята)'; 