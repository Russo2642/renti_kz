-- Откат уникального ограничения для payment_id

-- Удаляем уникальный индекс
DROP INDEX IF EXISTS idx_payments_payment_id_unique;

-- Удаляем уникальное ограничение
ALTER TABLE payments DROP CONSTRAINT IF EXISTS unique_payment_id;

-- Восстанавливаем обычный индекс
CREATE INDEX idx_payments_payment_id ON payments(payment_id); 