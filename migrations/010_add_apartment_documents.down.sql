-- Удаление триггера
DROP TRIGGER IF EXISTS update_apartment_documents_updated_at ON apartment_documents;

-- Удаление функции
DROP FUNCTION IF EXISTS update_apartment_documents_updated_at();

-- Удаление индекса
DROP INDEX IF EXISTS idx_apartment_documents_apartment_id;

-- Удаление таблицы
DROP TABLE IF EXISTS apartment_documents; 