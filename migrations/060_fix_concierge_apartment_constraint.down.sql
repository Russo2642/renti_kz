-- Откат исправления ограничения concierge_apartments

-- Удаляем новые ограничения
DROP INDEX IF EXISTS idx_unique_active_concierge_apartment;
ALTER TABLE concierge_apartments DROP CONSTRAINT IF EXISTS unique_concierge_apartment;

-- Восстанавливаем старое ограничение (с DEFERRABLE)
ALTER TABLE concierge_apartments ADD CONSTRAINT unique_active_concierge_apartment 
UNIQUE (concierge_id, apartment_id, is_active) DEFERRABLE INITIALLY DEFERRED;