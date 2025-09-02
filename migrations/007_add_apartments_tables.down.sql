-- Удаление триггеров
DROP TRIGGER IF EXISTS update_apartment_conditions_updated_at ON apartment_conditions;
DROP TRIGGER IF EXISTS update_apartment_locations_updated_at ON apartment_locations;
DROP TRIGGER IF EXISTS update_apartment_photos_updated_at ON apartment_photos;
DROP TRIGGER IF EXISTS update_apartments_updated_at ON apartments;

-- Удаление таблиц
DROP TABLE IF EXISTS apartment_locations;
DROP TABLE IF EXISTS apartment_photos;
DROP TABLE IF EXISTS apartments;
DROP TABLE IF EXISTS apartment_conditions; 