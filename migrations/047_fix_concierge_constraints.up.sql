-- Исправление ограничений в таблице консьержей
-- Пользователь может быть консьержем только одной активной квартиры
-- Квартира может иметь только одного активного консьержа

-- Удаляем старые ограничения
ALTER TABLE concierges DROP CONSTRAINT IF EXISTS unique_user_concierge;
ALTER TABLE concierges DROP CONSTRAINT IF EXISTS unique_apartment_concierge;

-- Создаем частичные уникальные индексы только для активных записей
CREATE UNIQUE INDEX unique_active_user_concierge 
ON concierges(user_id) 
WHERE is_active = TRUE;

CREATE UNIQUE INDEX unique_active_apartment_concierge 
ON concierges(apartment_id) 
WHERE is_active = TRUE;

-- Комментарии к новым ограничениям
COMMENT ON INDEX unique_active_user_concierge IS 'Пользователь может быть активным консьержем только одной квартиры';
COMMENT ON INDEX unique_active_apartment_concierge IS 'Квартира может иметь только одного активного консьержа'; 