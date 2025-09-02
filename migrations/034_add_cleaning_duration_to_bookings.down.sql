-- Удаление индекса
DROP INDEX IF EXISTS idx_bookings_cleaning_duration;

-- Удаление поля cleaning_duration из таблицы bookings
ALTER TABLE bookings DROP COLUMN IF EXISTS cleaning_duration; 