-- Удаляем индексы
DROP INDEX IF EXISTS idx_apartments_status_is_free;
DROP INDEX IF EXISTS idx_apartments_is_free;

-- Удаляем поле is_free
ALTER TABLE apartments DROP COLUMN IF EXISTS is_free; 