-- Откат частичных индексов для производительности

-- Удаляем все частичные индексы
DROP INDEX IF EXISTS idx_bookings_problematic_statuses;
DROP INDEX IF EXISTS idx_apartments_recently_updated;
DROP INDEX IF EXISTS idx_bookings_old_created;
DROP INDEX IF EXISTS idx_bookings_scheduler_tasks;