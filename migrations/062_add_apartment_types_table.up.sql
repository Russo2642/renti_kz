-- Создаем таблицу для типов квартир
CREATE TABLE apartment_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Заполняем базовые типы квартир
INSERT INTO apartment_types (name, description) VALUES
('Эконом', 'Базовые удобства, доступные цены'),
('Комфорт', 'Хорошие условия, оптимальное соотношение'),
('Бизнес', 'Повышенный комфорт для деловых поездок'),
('Премиум', 'Высокий уровень сервиса и удобств'),
('Люкс', 'Максимальный комфорт и эксклюзивность');

-- Добавляем поле apartment_type_id в таблицу apartments
ALTER TABLE apartments ADD COLUMN apartment_type_id INTEGER;

-- Добавляем foreign key constraint
ALTER TABLE apartments ADD CONSTRAINT fk_apartment_type 
    FOREIGN KEY (apartment_type_id) REFERENCES apartment_types(id);

-- Создаем индекс для быстрого поиска по типу
CREATE INDEX idx_apartments_type_id ON apartments(apartment_type_id);

-- Обновляем функцию updated_at для apartment_types
CREATE OR REPLACE FUNCTION update_apartment_types_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_apartment_types_updated_at 
    BEFORE UPDATE ON apartment_types 
    FOR EACH ROW EXECUTE FUNCTION update_apartment_types_updated_at_column();
