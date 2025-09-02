-- Удаляем старое ограничение с DEFERRABLE
ALTER TABLE cleaner_apartments DROP CONSTRAINT unique_active_cleaner_apartment;

-- Создаем частичный уникальный индекс для активных связей
CREATE UNIQUE INDEX unique_active_cleaner_apartment_idx 
    ON cleaner_apartments (cleaner_id, apartment_id) 
    WHERE is_active = TRUE;
