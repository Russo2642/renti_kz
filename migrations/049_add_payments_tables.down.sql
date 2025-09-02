-- Откат создания таблиц платежей

-- Удаляем триггер и функцию
DROP TRIGGER IF EXISTS trigger_payments_updated_at ON payments;
DROP FUNCTION IF EXISTS update_payments_updated_at();

-- Удаляем индексы (они удалятся автоматически с таблицами, но для явности)
DROP INDEX IF EXISTS idx_payment_logs_user_id;
DROP INDEX IF EXISTS idx_payment_logs_created_at;
DROP INDEX IF EXISTS idx_payment_logs_action;
DROP INDEX IF EXISTS idx_payment_logs_fp_payment_id;
DROP INDEX IF EXISTS idx_payment_logs_booking_id;
DROP INDEX IF EXISTS idx_payment_logs_payment_id;

DROP INDEX IF EXISTS idx_payments_processed_at;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_payment_id;
DROP INDEX IF EXISTS idx_payments_booking_id;

-- Удаляем таблицы (в правильном порядке из-за FK)
DROP TABLE IF EXISTS payment_logs;
DROP TABLE IF EXISTS payments; 