-- Откат добавления поля is_active в таблицу users

BEGIN;

-- Удаляем индекс
DROP INDEX IF EXISTS idx_users_is_active;

-- Удаляем поле is_active
ALTER TABLE users 
DROP COLUMN IF EXISTS is_active;

COMMIT; 