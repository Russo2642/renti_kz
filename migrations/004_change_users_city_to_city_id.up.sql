-- Добавление временного поля city_id в таблицу users
ALTER TABLE users ADD COLUMN city_id INT;

-- Добавление внешнего ключа
ALTER TABLE users ADD CONSTRAINT fk_users_city_id FOREIGN KEY (city_id) REFERENCES cities(id);

-- Создание индекса
CREATE INDEX IF NOT EXISTS idx_users_city_id ON users(city_id); 