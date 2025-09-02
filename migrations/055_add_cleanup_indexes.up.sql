-- Добавление индексов для оптимизации операций очистки

-- Индекс для очистки просроченных бронирований
-- Покрывает WHERE status IN ('created', 'awaiting_payment', 'pending') AND start_date < NOW()
CREATE INDEX idx_bookings_cleanup 
ON bookings(status, start_date) 
WHERE status IN ('created', 'awaiting_payment', 'pending');

-- Индекс для очистки просроченных продлений  
-- Покрывает WHERE status = 'awaiting_payment' AND created_at < NOW() - INTERVAL '2 hours'
CREATE INDEX idx_booking_extensions_cleanup
ON booking_extensions(status, created_at)
WHERE status = 'awaiting_payment';

-- Составной индекс для оптимизации CheckApartmentAvailability
-- Улучшает производительность проверки доступности квартир
CREATE INDEX idx_bookings_availability_optimized
ON bookings(apartment_id, status, start_date, end_date)
WHERE status IN ('pending', 'approved', 'active');

-- Индекс для работы с cleaning_duration в проверке доступности
-- Покрывает (end_date + INTERVAL '1 minute' * cleaning_duration) условия
CREATE INDEX idx_bookings_end_date_cleaning
ON bookings(apartment_id, end_date, cleaning_duration, status)
WHERE status IN ('pending', 'approved', 'active');