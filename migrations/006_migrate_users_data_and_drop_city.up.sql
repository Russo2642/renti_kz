-- Обновление поля city_id для существующих пользователей
UPDATE users SET city_id = 
    CASE city
        WHEN 'Алматы' THEN (SELECT id FROM cities WHERE name = 'Алматы')
        WHEN 'Астана' THEN (SELECT id FROM cities WHERE name = 'Астана')
        WHEN 'Шымкент' THEN (SELECT id FROM cities WHERE name = 'Шымкент')
        WHEN 'Караганда' THEN (SELECT id FROM cities WHERE name = 'Караганда')
        WHEN 'Актобе' THEN (SELECT id FROM cities WHERE name = 'Актобе')
        WHEN 'Тараз' THEN (SELECT id FROM cities WHERE name = 'Тараз')
        WHEN 'Павлодар' THEN (SELECT id FROM cities WHERE name = 'Павлодар')
        WHEN 'Усть-Каменогорск' THEN (SELECT id FROM cities WHERE name = 'Усть-Каменогорск')
        WHEN 'Семей' THEN (SELECT id FROM cities WHERE name = 'Семей')
        WHEN 'Атырау' THEN (SELECT id FROM cities WHERE name = 'Атырау')
        WHEN 'Костанай' THEN (SELECT id FROM cities WHERE name = 'Костанай')
        WHEN 'Кызылорда' THEN (SELECT id FROM cities WHERE name = 'Кызылорда')
        WHEN 'Уральск' THEN (SELECT id FROM cities WHERE name = 'Уральск')
        WHEN 'Петропавловск' THEN (SELECT id FROM cities WHERE name = 'Петропавловск')
        WHEN 'Кокшетау' THEN (SELECT id FROM cities WHERE name = 'Кокшетау')
        WHEN 'Талдыкорган' THEN (SELECT id FROM cities WHERE name = 'Талдыкорган')
        WHEN 'Экибастуз' THEN (SELECT id FROM cities WHERE name = 'Экибастуз')
        WHEN 'Туркестан' THEN (SELECT id FROM cities WHERE name = 'Туркестан')
        ELSE (SELECT id FROM cities WHERE name = 'Алматы') -- Значение по умолчанию
    END
WHERE city_id IS NULL;

-- Сделать поле city_id NOT NULL и удалить поле city
ALTER TABLE users ALTER COLUMN city_id SET NOT NULL;
ALTER TABLE users DROP COLUMN city; 