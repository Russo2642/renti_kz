-- Удаление индексов для типов аренды
DROP INDEX IF EXISTS idx_apartments_rental_types;
DROP INDEX IF EXISTS idx_apartments_rental_type_daily;
DROP INDEX IF EXISTS idx_apartments_rental_type_hourly;

-- Удаление ограничения
ALTER TABLE apartments DROP CONSTRAINT IF EXISTS check_rental_type_selected;

-- Удаление колонок типов аренды
ALTER TABLE apartments DROP COLUMN IF EXISTS rental_type_daily;
ALTER TABLE apartments DROP COLUMN IF EXISTS rental_type_hourly; 