-- Удаление индексов
DROP INDEX IF EXISTS idx_microdistricts_district_id;
DROP INDEX IF EXISTS idx_districts_city_id;
DROP INDEX IF EXISTS idx_cities_region_id;

-- Удаление триггеров
DROP TRIGGER IF EXISTS update_microdistricts_updated_at ON microdistricts;
DROP TRIGGER IF EXISTS update_districts_updated_at ON districts;
DROP TRIGGER IF EXISTS update_cities_updated_at ON cities;
DROP TRIGGER IF EXISTS update_regions_updated_at ON regions;

-- Удаление таблиц
DROP TABLE IF EXISTS microdistricts;
DROP TABLE IF EXISTS districts;
DROP TABLE IF EXISTS cities;
DROP TABLE IF EXISTS regions; 