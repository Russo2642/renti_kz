-- Удаляем добавленные колонки из таблицы apartments
ALTER TABLE apartments DROP COLUMN IF EXISTS description;
ALTER TABLE apartments DROP COLUMN IF EXISTS house_rules_id;
ALTER TABLE apartments DROP COLUMN IF EXISTS amenities_id;

-- Удаляем таблицы
DROP TABLE IF EXISTS popular_amenities;
DROP TABLE IF EXISTS house_rules; 