-- Удаление индекса
DROP INDEX IF EXISTS idx_users_city_id;

-- Удаление внешнего ключа
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_city_id;

-- Удаление поля city_id
ALTER TABLE users DROP COLUMN IF EXISTS city_id; 