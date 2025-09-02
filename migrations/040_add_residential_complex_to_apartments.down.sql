-- Удаление индекса и поля residential_complex из таблицы apartments
DROP INDEX IF EXISTS idx_apartments_residential_complex;
ALTER TABLE apartments DROP COLUMN IF EXISTS residential_complex; 