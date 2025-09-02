-- Добавление поля residential_complex в таблицу apartments
ALTER TABLE apartments 
ADD COLUMN residential_complex VARCHAR(255);
 
-- Создание индекса для поля residential_complex для быстрого поиска
CREATE INDEX idx_apartments_residential_complex ON apartments(residential_complex); 