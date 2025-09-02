-- Восстанавливаем старые колонки
ALTER TABLE apartments ADD COLUMN house_rules_id INTEGER REFERENCES house_rules(id);
ALTER TABLE apartments ADD COLUMN amenities_id INTEGER REFERENCES popular_amenities(id);

-- Переносим данные обратно (берем первое правило и первое удобство из связующих таблиц)
UPDATE apartments a
SET house_rules_id = (
    SELECT house_rule_id 
    FROM apartment_house_rules 
    WHERE apartment_id = a.id 
    ORDER BY id 
    LIMIT 1
);

UPDATE apartments a
SET amenities_id = (
    SELECT amenity_id 
    FROM apartment_amenities 
    WHERE apartment_id = a.id 
    ORDER BY id 
    LIMIT 1
);

-- Изменяем тип apartment_number с integer обратно на varchar
-- Сначала сохраняем данные во временной колонке
ALTER TABLE apartments ADD COLUMN apartment_number_varchar VARCHAR(50);
UPDATE apartments SET apartment_number_varchar = apartment_number::text;

-- Удаляем старую колонку и переименовываем новую
ALTER TABLE apartments DROP COLUMN apartment_number;
ALTER TABLE apartments RENAME COLUMN apartment_number_varchar TO apartment_number;

-- Удаляем связующие таблицы
DROP TABLE IF EXISTS apartment_house_rules;
DROP TABLE IF EXISTS apartment_amenities; 