-- Удаление таблицы консьержей
DROP TRIGGER IF EXISTS update_concierges_updated_at ON concierges;
DROP TABLE IF EXISTS concierges; 