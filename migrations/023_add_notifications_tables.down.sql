-- Удаляем триггеры
DROP TRIGGER IF EXISTS update_user_devices_updated_at ON user_devices;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаляем таблицы
DROP TABLE IF EXISTS user_devices;
DROP TABLE IF EXISTS notifications;

-- Удаляем типы
DROP TYPE IF EXISTS device_type;
DROP TYPE IF EXISTS notification_priority;
DROP TYPE IF EXISTS notification_type; 