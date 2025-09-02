-- Откат изменений - делаем apartment_id обязательным снова
-- Сначала устанавливаем значение по умолчанию для NULL записей

UPDATE locks SET apartment_id = 0 WHERE apartment_id IS NULL;

ALTER TABLE locks 
ALTER COLUMN apartment_id SET NOT NULL;

-- Возвращаем старый комментарий
COMMENT ON COLUMN locks.apartment_id IS 'ID квартиры'; 