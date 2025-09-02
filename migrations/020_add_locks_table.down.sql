-- Удаление связей и индексов
DROP INDEX IF EXISTS idx_apartments_lock_id;
ALTER TABLE apartments DROP COLUMN IF EXISTS lock_id;

-- Удаление триггеров и функций
DROP TRIGGER IF EXISTS log_lock_status_change_trigger ON locks;
DROP FUNCTION IF EXISTS log_lock_status_change();

-- Удаление индексов
DROP INDEX IF EXISTS idx_lock_status_logs_source;
DROP INDEX IF EXISTS idx_lock_status_logs_created_at;
DROP INDEX IF EXISTS idx_lock_status_logs_lock_id;
DROP INDEX IF EXISTS idx_locks_heartbeat;
DROP INDEX IF EXISTS idx_locks_online;
DROP INDEX IF EXISTS idx_locks_status;
DROP INDEX IF EXISTS idx_locks_unique_id;
DROP INDEX IF EXISTS idx_locks_apartment_id;

-- Удаление таблиц
DROP TABLE IF EXISTS lock_status_logs;
DROP TABLE IF EXISTS locks; 