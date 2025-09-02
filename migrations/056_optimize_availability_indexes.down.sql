-- Откат оптимизации индексов для системы is_free

-- Удаляем созданные индексы
DROP INDEX IF EXISTS idx_bookings_apartment_status_dates;
DROP INDEX IF EXISTS idx_bookings_created_status_time;
DROP INDEX IF EXISTS idx_bookings_active_timeframe;
DROP INDEX IF EXISTS idx_apartments_is_free_with_status;
DROP INDEX IF EXISTS idx_bookings_near_future;

-- Восстанавливаем простой индекс apartments.is_free
CREATE INDEX idx_apartments_is_free ON apartments(is_free);