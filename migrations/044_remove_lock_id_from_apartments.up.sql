-- Удаление избыточного поля lock_id из таблицы apartments
-- Связь с замками будет поддерживаться через locks.apartment_id

-- Удаляем индекс для lock_id
DROP INDEX IF EXISTS idx_apartments_lock_id;

-- Удаляем внешний ключ constraint если есть
ALTER TABLE apartments DROP CONSTRAINT IF EXISTS fk_apartments_lock_id;

-- Удаляем колонку lock_id
ALTER TABLE apartments DROP COLUMN IF EXISTS lock_id;

-- Комментарий
COMMENT ON TABLE apartments IS 'Квартиры без дублирования связи с замками - используется locks.apartment_id'; 