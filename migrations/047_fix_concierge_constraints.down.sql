-- Откат изменений ограничений в таблице консьержей

-- Удаляем новые частичные индексы
DROP INDEX IF EXISTS unique_active_user_concierge;
DROP INDEX IF EXISTS unique_active_apartment_concierge;

-- Восстанавливаем старые ограничения
ALTER TABLE concierges 
ADD CONSTRAINT unique_user_concierge UNIQUE(user_id),
ADD CONSTRAINT unique_apartment_concierge UNIQUE(apartment_id); 