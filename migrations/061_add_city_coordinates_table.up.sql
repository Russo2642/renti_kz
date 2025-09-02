-- Создание таблицы координат городов
CREATE TABLE IF NOT EXISTS city_coordinates (
    id SERIAL PRIMARY KEY,
    city_id INT NOT NULL UNIQUE,
    latitude DECIMAL(10, 7) NOT NULL,
    longitude DECIMAL(11, 7) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (city_id) REFERENCES cities(id) ON DELETE CASCADE
);

-- Добавление триггера для обновления updated_at
CREATE TRIGGER update_city_coordinates_updated_at
BEFORE UPDATE ON city_coordinates
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Заполнение координат для существующих городов
INSERT INTO city_coordinates (city_id, latitude, longitude) VALUES
(1, 43.2389, 76.8897),   -- Алматы
(2, 51.1605, 71.4704),   -- Астана
(3, 42.3417, 69.5901),   -- Шымкент
(4, 49.8028, 73.1022),   -- Караганда
(5, 50.2839, 57.1670),   -- Актобе
(6, 42.8991, 71.3658),   -- Тараз
(7, 52.2870, 76.9674),   -- Павлодар
(8, 49.9577, 82.6111),   -- Усть-Каменогорск
(9, 50.4110, 80.2270),   -- Семей
(10, 47.1164, 51.9250),  -- Атырау
(11, 53.2144, 63.6246),  -- Костанай
(12, 44.8488, 65.4823),  -- Кызылорда
(13, 51.2224, 51.3724),  -- Уральск
(14, 54.8730, 69.1430),  -- Петропавловск
(15, 53.2888, 69.3902),  -- Кокшетау
(16, 45.0156, 78.3731),  -- Талдыкорган
(17, 51.7231, 75.3224),  -- Экибастуз
(18, 43.2973, 68.2534)   -- Туркестан
ON CONFLICT (city_id) DO NOTHING;
