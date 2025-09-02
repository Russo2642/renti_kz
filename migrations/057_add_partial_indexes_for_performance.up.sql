-- Дополнительные частичные индексы для максимальной производительности

-- 1. Частичный индекс только для проблемных статусов бронирований
-- Значительно уменьшает размер индекса и ускоряет запросы
CREATE INDEX IF NOT EXISTS idx_bookings_problematic_statuses 
ON bookings(apartment_id, start_date, end_date, created_at) 
WHERE status IN ('created', 'pending', 'awaiting_payment', 'approved');

-- 2. Индекс для быстрого поиска "горячих" квартир (недавно изменённых)
CREATE INDEX IF NOT EXISTS idx_apartments_recently_updated 
ON apartments(updated_at, id) 
WHERE is_free = false;

-- 3. Индекс для очистки старых created бронирований
-- Убираем NOW() из WHERE так как функции в предикатах индексов должны быть IMMUTABLE
CREATE INDEX IF NOT EXISTS idx_bookings_old_created 
ON bookings(created_at, apartment_id) 
WHERE status = 'created';

-- 4. Составной индекс для шедулера (активация/завершение бронирований)
CREATE INDEX IF NOT EXISTS idx_bookings_scheduler_tasks 
ON bookings(status, start_date, end_date, apartment_id) 
WHERE status IN ('approved', 'active');

-- Принудительно обновляем статистику для планировщика запросов
ANALYZE bookings;
ANALYZE apartments;