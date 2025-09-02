-- Удаление индекса для цены за сутки
DROP INDEX IF EXISTS idx_apartments_daily_price;

-- Удаление ограничения
ALTER TABLE apartments DROP CONSTRAINT IF EXISTS check_daily_price_non_negative;

-- Удаление колонки daily_price
ALTER TABLE apartments DROP COLUMN IF EXISTS daily_price; 