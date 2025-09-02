-- Удаление колонки service_fee_percentage из таблицы apartments

-- Удаляем индекс
DROP INDEX IF EXISTS idx_apartments_service_fee_percentage;

-- Удаляем ограничение
ALTER TABLE apartments DROP CONSTRAINT IF EXISTS check_service_fee_percentage_range;

-- Удаляем колонку
ALTER TABLE apartments DROP COLUMN IF EXISTS service_fee_percentage;