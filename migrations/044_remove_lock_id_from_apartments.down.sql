-- Rollback: восстановление поля lock_id в таблице apartments

-- Добавляем колонку lock_id обратно
ALTER TABLE apartments ADD COLUMN lock_id INTEGER;

-- Добавляем внешний ключ
ALTER TABLE apartments ADD CONSTRAINT fk_apartments_lock_id 
FOREIGN KEY (lock_id) REFERENCES locks(id) ON DELETE SET NULL;

-- Восстанавливаем индекс
CREATE INDEX idx_apartments_lock_id ON apartments(lock_id);

-- Синхронизируем данные: устанавливаем lock_id на основе locks.apartment_id
UPDATE apartments 
SET lock_id = (
    SELECT l.id 
    FROM locks l 
    WHERE l.apartment_id = apartments.id 
    LIMIT 1
);

-- Комментарий
COMMENT ON COLUMN apartments.lock_id IS 'Ссылка на основной замок квартиры (восстановлено)'; 