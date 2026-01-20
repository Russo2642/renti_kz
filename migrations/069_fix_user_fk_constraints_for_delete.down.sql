-- Откат: возвращаем constraints без ON DELETE SET NULL

-- 1. Откат lock_status_logs.user_id
ALTER TABLE lock_status_logs 
DROP CONSTRAINT IF EXISTS lock_status_logs_user_id_fkey;

ALTER TABLE lock_status_logs
ADD CONSTRAINT lock_status_logs_user_id_fkey 
FOREIGN KEY (user_id) REFERENCES users(id);

-- 2. Откат payment_logs.user_id
ALTER TABLE payment_logs 
DROP CONSTRAINT IF EXISTS payment_logs_user_id_fkey;

ALTER TABLE payment_logs
ADD CONSTRAINT payment_logs_user_id_fkey 
FOREIGN KEY (user_id) REFERENCES users(id);

