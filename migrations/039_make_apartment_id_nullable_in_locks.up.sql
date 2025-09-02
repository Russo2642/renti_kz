-- Изменяем apartment_id на nullable в таблице locks
-- Это позволит создавать замки без привязки к квартире

ALTER TABLE locks 
ALTER COLUMN apartment_id DROP NOT NULL;

-- Добавляем комментарий для разъяснения изменения
COMMENT ON COLUMN locks.apartment_id IS 'ID квартиры (может быть NULL если замок не привязан)'; 