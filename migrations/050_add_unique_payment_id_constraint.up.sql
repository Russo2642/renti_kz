-- Добавление уникального ограничения для предотвращения дублирования payment_id

-- Сначала проверим и удалим дубликаты если есть (на всякий случай)
-- Оставляем только самые ранние записи для каждого payment_id
DELETE FROM payments 
WHERE id NOT IN (
    SELECT MIN(id) 
    FROM payments 
    GROUP BY payment_id
);

-- Добавляем уникальное ограничение на payment_id
-- (на всякий случай проверяем что его еще нет)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'unique_payment_id' 
        AND conrelid = 'payments'::regclass
    ) THEN
        ALTER TABLE payments 
        ADD CONSTRAINT unique_payment_id UNIQUE (payment_id);
    END IF;
END $$;

-- Обновляем существующий индекс чтобы он был уникальным
DROP INDEX IF EXISTS idx_payments_payment_id;
CREATE UNIQUE INDEX idx_payments_payment_id_unique ON payments(payment_id);

-- Комментарий для документации
COMMENT ON CONSTRAINT unique_payment_id ON payments IS 'Предотвращает использование одного payment_id для нескольких бронирований'; 