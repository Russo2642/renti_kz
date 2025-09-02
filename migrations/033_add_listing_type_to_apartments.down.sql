-- Удаление индекса
DROP INDEX IF EXISTS idx_apartments_listing_type;

-- Удаление поля listing_type из таблицы apartments
ALTER TABLE apartments DROP COLUMN IF EXISTS listing_type; 