-- Добавление поля city
ALTER TABLE users ADD COLUMN city VARCHAR(100);

-- Обновление поля city на основе city_id
UPDATE users SET city = 
    (SELECT name FROM cities WHERE id = users.city_id)
WHERE city IS NULL;

-- Установка поля city как NOT NULL
ALTER TABLE users ALTER COLUMN city SET NOT NULL; 