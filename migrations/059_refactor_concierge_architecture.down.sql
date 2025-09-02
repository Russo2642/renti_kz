-- Откат рефакторинга архитектуры консьержей

-- Удаляем триггер и функцию
DROP TRIGGER IF EXISTS trigger_update_concierge_apartments_updated_at ON concierge_apartments;
DROP FUNCTION IF EXISTS update_concierge_apartments_updated_at();

-- Удаляем индексы
DROP INDEX IF EXISTS idx_concierge_apartments_assigned_at;
DROP INDEX IF EXISTS idx_concierge_apartments_active;
DROP INDEX IF EXISTS idx_concierge_apartments_apartment_id;
DROP INDEX IF EXISTS idx_concierge_apartments_concierge_id;

-- Удаляем промежуточную таблицу
DROP TABLE IF EXISTS concierge_apartments CASCADE;

-- Восстанавливаем колонку apartment_id в таблице concierges
ALTER TABLE concierges ADD COLUMN apartment_id INTEGER REFERENCES apartments(id) ON DELETE CASCADE;

-- Создаем индекс для apartment_id
CREATE INDEX idx_concierges_apartment_id ON concierges(apartment_id);