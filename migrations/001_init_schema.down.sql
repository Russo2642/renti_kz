-- Удаление триггеров
DROP TRIGGER IF EXISTS update_renters_updated_at ON renters;
DROP TRIGGER IF EXISTS update_property_owners_updated_at ON property_owners;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Удаление функции для обновления updated_at
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаление индексов
DROP INDEX IF EXISTS idx_renters_user_id;
DROP INDEX IF EXISTS idx_property_owners_user_id;
DROP INDEX IF EXISTS idx_users_profile_status;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_phone;

-- Удаление таблиц
DROP TABLE IF EXISTS renters;
DROP TABLE IF EXISTS property_owners;
DROP TABLE IF EXISTS users; 