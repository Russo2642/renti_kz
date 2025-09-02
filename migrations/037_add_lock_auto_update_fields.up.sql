-- Добавляем поля для автоматического обновления статуса и батареи замков
ALTER TABLE locks 
-- Расширенная информация о батарее
ADD COLUMN battery_type VARCHAR(20) NOT NULL DEFAULT 'unknown' CHECK (battery_type IN ('alkaline', 'lithium', 'unknown')),
ADD COLUMN charging_status VARCHAR(20) CHECK (charging_status IN ('not_charging', 'charging', 'full')),
ADD COLUMN last_battery_check TIMESTAMP WITH TIME ZONE,

-- Поля для интеграции с Tuya
ADD COLUMN last_tuya_sync TIMESTAMP WITH TIME ZONE,
ADD COLUMN auto_update_enabled BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN webhook_configured BOOLEAN NOT NULL DEFAULT false;

-- Обновляем список допустимых источников изменений в lock_status_logs
ALTER TABLE lock_status_logs DROP CONSTRAINT lock_status_logs_change_source_check;
ALTER TABLE lock_status_logs ADD CONSTRAINT lock_status_logs_change_source_check 
CHECK (change_source IN ('api', 'manual', 'system', 'tuya', 'webhook'));

-- Создаем индексы для оптимизации
CREATE INDEX idx_locks_auto_update ON locks(auto_update_enabled);
CREATE INDEX idx_locks_tuya_sync ON locks(last_tuya_sync);
CREATE INDEX idx_locks_battery_type ON locks(battery_type);
CREATE INDEX idx_locks_webhook_configured ON locks(webhook_configured);

-- Комментарии для новых полей
COMMENT ON COLUMN locks.battery_type IS 'Тип батареи: alkaline (щелочная), lithium (литий-ионная), unknown (неизвестный)';
COMMENT ON COLUMN locks.charging_status IS 'Статус зарядки для литий-ионных батарей';
COMMENT ON COLUMN locks.last_battery_check IS 'Время последней проверки батареи';
COMMENT ON COLUMN locks.last_tuya_sync IS 'Время последней синхронизации с Tuya API';
COMMENT ON COLUMN locks.auto_update_enabled IS 'Включено ли автоматическое обновление статуса и батареи';
COMMENT ON COLUMN locks.webhook_configured IS 'Настроены ли webhooks для автоматического получения событий'; 