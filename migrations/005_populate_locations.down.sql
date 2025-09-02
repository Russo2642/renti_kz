-- Удаление данных из таблицы районов
DELETE FROM districts WHERE city_id IN (SELECT id FROM cities WHERE name = 'Алматы');

-- Удаление данных из таблицы городов
DELETE FROM cities;

-- Удаление данных из таблицы регионов
DELETE FROM regions; 