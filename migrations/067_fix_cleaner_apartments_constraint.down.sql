-- Удаляем частичный индекс
DROP INDEX IF EXISTS unique_active_cleaner_apartment_idx;

-- Восстанавливаем старое ограничение с DEFERRABLE
ALTER TABLE cleaner_apartments ADD CONSTRAINT unique_active_cleaner_apartment 
    UNIQUE (cleaner_id, apartment_id, is_active) 
    DEFERRABLE INITIALLY DEFERRED;
