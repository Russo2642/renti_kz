-- Создание таблицы состояний квартир
CREATE TABLE IF NOT EXISTS apartment_conditions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Вставка начальных данных в таблицу состояний квартир
INSERT INTO apartment_conditions (name, description) VALUES
    ('Свежий ремонт, новая мебель', 'Квартира с новым ремонтом и новой мебелью'),
    ('Не новый, но аккуратный и чистый', 'Квартира в хорошем состоянии, чистая и аккуратная'),
    ('Без ремонта', 'Квартира требует ремонта или находится в базовом состоянии');

-- Создание таблицы квартир
CREATE TABLE IF NOT EXISTS apartments (
    id SERIAL PRIMARY KEY,
    owner_id INT NOT NULL REFERENCES property_owners(id) ON DELETE CASCADE,
    city_id INT NOT NULL REFERENCES cities(id),
    district_id INT NOT NULL REFERENCES districts(id),
    microdistrict_id INT REFERENCES microdistricts(id),
    street VARCHAR(255) NOT NULL,
    building VARCHAR(50) NOT NULL,
    apartment_number VARCHAR(50) NOT NULL,
    room_count INT NOT NULL,
    total_area DECIMAL(10, 2) NOT NULL,
    kitchen_area DECIMAL(10, 2) NOT NULL,
    floor INT NOT NULL,
    total_floors INT NOT NULL,
    condition_id INT NOT NULL REFERENCES apartment_conditions(id),
    price INT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    moderator_comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_owner FOREIGN KEY (owner_id) REFERENCES property_owners(id) ON DELETE CASCADE,
    CONSTRAINT fk_city FOREIGN KEY (city_id) REFERENCES cities(id),
    CONSTRAINT fk_district FOREIGN KEY (district_id) REFERENCES districts(id),
    CONSTRAINT fk_microdistrict FOREIGN KEY (microdistrict_id) REFERENCES microdistricts(id),
    CONSTRAINT fk_condition FOREIGN KEY (condition_id) REFERENCES apartment_conditions(id)
);

-- Создание индексов для таблицы квартир
CREATE INDEX idx_apartments_owner_id ON apartments(owner_id);
CREATE INDEX idx_apartments_city_id ON apartments(city_id);
CREATE INDEX idx_apartments_district_id ON apartments(district_id);
CREATE INDEX idx_apartments_status ON apartments(status);

-- Создание таблицы фотографий квартир
CREATE TABLE IF NOT EXISTS apartment_photos (
    id SERIAL PRIMARY KEY,
    apartment_id INT NOT NULL,
    url VARCHAR(500) NOT NULL,
    "order" INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_apartment FOREIGN KEY (apartment_id) REFERENCES apartments(id) ON DELETE CASCADE
);

-- Создание индекса для таблицы фотографий
CREATE INDEX idx_apartment_photos_apartment_id ON apartment_photos(apartment_id);

-- Создание таблицы координат квартир
CREATE TABLE IF NOT EXISTS apartment_locations (
    id SERIAL PRIMARY KEY,
    apartment_id INT NOT NULL UNIQUE,
    latitude DECIMAL(11, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_apartment FOREIGN KEY (apartment_id) REFERENCES apartments(id) ON DELETE CASCADE
);

-- Создание индекса для таблицы координат
CREATE INDEX idx_apartment_locations_apartment_id ON apartment_locations(apartment_id);

-- Добавление триггера для автоматического обновления updated_at при изменении записей
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

-- Добавление триггера для таблицы квартир
CREATE TRIGGER update_apartments_updated_at
BEFORE UPDATE ON apartments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Добавление триггера для таблицы фотографий
CREATE TRIGGER update_apartment_photos_updated_at
BEFORE UPDATE ON apartment_photos
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Добавление триггера для таблицы координат
CREATE TRIGGER update_apartment_locations_updated_at
BEFORE UPDATE ON apartment_locations
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Добавление триггера для таблицы состояний
CREATE TRIGGER update_apartment_conditions_updated_at
BEFORE UPDATE ON apartment_conditions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column(); 