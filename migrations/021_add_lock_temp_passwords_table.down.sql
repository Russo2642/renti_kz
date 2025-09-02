-- Удаляем триггер и функцию
DROP TRIGGER IF EXISTS trigger_lock_temp_passwords_updated_at ON lock_temp_passwords;
DROP FUNCTION IF EXISTS update_lock_temp_passwords_updated_at();

-- Удаляем таблицу временных паролей
DROP TABLE IF EXISTS lock_temp_passwords;

-- Удаляем новые поля из таблицы замков
ALTER TABLE locks 
DROP COLUMN IF EXISTS tuya_device_id,
DROP COLUMN IF EXISTS owner_password; 