-- Исправление ограничения concierge_apartments

-- Удаляем проблемное ограничение с DEFERRABLE
ALTER TABLE concierge_apartments DROP CONSTRAINT IF EXISTS unique_active_concierge_apartment;

-- Создаем простое уникальное ограничение без DEFERRABLE
ALTER TABLE concierge_apartments ADD CONSTRAINT unique_concierge_apartment UNIQUE (concierge_id, apartment_id);

-- Создаем частичный индекс для активных записей (более эффективно)
CREATE UNIQUE INDEX idx_unique_active_concierge_apartment 
ON concierge_apartments (concierge_id, apartment_id) 
WHERE is_active = true;