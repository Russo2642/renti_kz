-- Откат добавления индексов для очистки

-- Удаление индекса для очистки просроченных бронирований
DROP INDEX IF EXISTS idx_bookings_cleanup;

-- Удаление индекса для очистки просроченных продлений
DROP INDEX IF EXISTS idx_booking_extensions_cleanup;

-- Удаление составного индекса для CheckApartmentAvailability
DROP INDEX IF EXISTS idx_bookings_availability_optimized;

-- Удаление индекса для работы с cleaning_duration
DROP INDEX IF EXISTS idx_bookings_end_date_cleaning;