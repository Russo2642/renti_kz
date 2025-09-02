-- Отменяем добавление полей для автоматического обновления замков

-- Удаляем индексы
DROP INDEX IF EXISTS idx_locks_webhook_configured;
DROP INDEX IF EXISTS idx_locks_battery_type;
DROP INDEX IF EXISTS idx_locks_tuya_sync;
DROP INDEX IF EXISTS idx_locks_auto_update;

-- Возвращаем старое ограничение для change_source
ALTER TABLE lock_status_logs DROP CONSTRAINT lock_status_logs_change_source_check;
ALTER TABLE lock_status_logs ADD CONSTRAINT lock_status_logs_change_source_check 
CHECK (change_source IN ('api', 'manual', 'system'));

-- Удаляем добавленные поля
ALTER TABLE locks 
DROP COLUMN webhook_configured,
DROP COLUMN auto_update_enabled,
DROP COLUMN last_tuya_sync,
DROP COLUMN last_battery_check,
DROP COLUMN charging_status,
DROP COLUMN battery_type; 