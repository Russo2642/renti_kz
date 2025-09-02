-- Удаляем trigger и функцию
DROP TRIGGER IF EXISTS update_apartment_types_updated_at ON apartment_types;
DROP FUNCTION IF EXISTS update_apartment_types_updated_at_column();

-- Удаляем индексы
DROP INDEX IF EXISTS idx_apartments_type_id;

-- Удаляем foreign key constraint
ALTER TABLE apartments DROP CONSTRAINT IF EXISTS fk_apartment_type;

-- Удаляем поле из apartments
ALTER TABLE apartments DROP COLUMN IF EXISTS apartment_type_id;

-- Удаляем таблицу apartment_types
DROP TABLE IF EXISTS apartment_types;
