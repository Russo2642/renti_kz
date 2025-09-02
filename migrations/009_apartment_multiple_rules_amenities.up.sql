-- Создаем связующие таблицы для множественных связей
CREATE TABLE IF NOT EXISTS apartment_house_rules (
    id SERIAL PRIMARY KEY,
    apartment_id INTEGER NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    house_rule_id INTEGER NOT NULL REFERENCES house_rules(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(apartment_id, house_rule_id)
);

CREATE TABLE IF NOT EXISTS apartment_amenities (
    id SERIAL PRIMARY KEY,
    apartment_id INTEGER NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    amenity_id INTEGER NOT NULL REFERENCES popular_amenities(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(apartment_id, amenity_id)
);

-- Создаем индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_apartment_house_rules_apartment_id ON apartment_house_rules(apartment_id);
CREATE INDEX IF NOT EXISTS idx_apartment_house_rules_house_rule_id ON apartment_house_rules(house_rule_id);
CREATE INDEX IF NOT EXISTS idx_apartment_amenities_apartment_id ON apartment_amenities(apartment_id);
CREATE INDEX IF NOT EXISTS idx_apartment_amenities_amenity_id ON apartment_amenities(amenity_id);

-- Копируем данные из старых полей в новые связующие таблицы
INSERT INTO apartment_house_rules (apartment_id, house_rule_id)
SELECT id, house_rules_id FROM apartments
WHERE house_rules_id IS NOT NULL;

INSERT INTO apartment_amenities (apartment_id, amenity_id)
SELECT id, amenities_id FROM apartments 
WHERE amenities_id IS NOT NULL;

-- Удаляем старые колонки
ALTER TABLE apartments DROP COLUMN IF EXISTS house_rules_id;
ALTER TABLE apartments DROP COLUMN IF EXISTS amenities_id;

-- Изменяем тип apartment_number с varchar на integer
-- Сначала сохраняем данные во временной колонке
ALTER TABLE apartments ADD COLUMN apartment_number_int INTEGER;
UPDATE apartments SET apartment_number_int = apartment_number::integer WHERE apartment_number ~ '^[0-9]+$';

-- Удаляем старую колонку и переименовываем новую
ALTER TABLE apartments DROP COLUMN apartment_number;
ALTER TABLE apartments RENAME COLUMN apartment_number_int TO apartment_number; 