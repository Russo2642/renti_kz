-- Создание таблицы избранных квартир
CREATE TABLE IF NOT EXISTS favorites (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    apartment_id INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT fk_favorites_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_favorites_apartment FOREIGN KEY (apartment_id) REFERENCES apartments(id) ON DELETE CASCADE,
    CONSTRAINT uk_favorites_user_apartment UNIQUE (user_id, apartment_id)
);

-- Создание индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_favorites_user_id ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_apartment_id ON favorites(apartment_id);
CREATE INDEX IF NOT EXISTS idx_favorites_created_at ON favorites(created_at);

-- Добавление триггера для автоматического обновления updated_at
CREATE TRIGGER update_favorites_updated_at
BEFORE UPDATE ON favorites
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Добавление комментариев к таблице и колонкам
COMMENT ON TABLE favorites IS 'Таблица избранных квартир пользователей';
COMMENT ON COLUMN favorites.id IS 'Уникальный идентификатор записи';
COMMENT ON COLUMN favorites.user_id IS 'Идентификатор пользователя';
COMMENT ON COLUMN favorites.apartment_id IS 'Идентификатор квартиры';
COMMENT ON COLUMN favorites.created_at IS 'Время добавления в избранное';
COMMENT ON COLUMN favorites.updated_at IS 'Время последнего обновления'; 