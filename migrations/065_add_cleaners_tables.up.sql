-- Создание таблицы уборщиц
CREATE TABLE cleaners (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    schedule JSONB, -- расписание работы уборщицы
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Уникальное ограничение: пользователь может быть только одной активной уборщицей
    CONSTRAINT unique_active_user_cleaner UNIQUE(user_id)
);

-- Создание таблицы связи уборщиц с квартирами (many-to-many)
CREATE TABLE cleaner_apartments (
    id SERIAL PRIMARY KEY,
    cleaner_id INTEGER NOT NULL REFERENCES cleaners(id) ON DELETE CASCADE,
    apartment_id INTEGER NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Уникальное ограничение: одна квартира может иметь несколько уборщиц, но без дублирования активных связей
    CONSTRAINT unique_active_cleaner_apartment UNIQUE (cleaner_id, apartment_id, is_active) 
        DEFERRABLE INITIALLY DEFERRED
);

-- Индексы для оптимизации запросов cleaners
CREATE INDEX idx_cleaners_user_id ON cleaners(user_id);
CREATE INDEX idx_cleaners_active ON cleaners(is_active) WHERE is_active = TRUE;

-- Индексы для оптимизации запросов cleaner_apartments
CREATE INDEX idx_cleaner_apartments_cleaner_id ON cleaner_apartments(cleaner_id);
CREATE INDEX idx_cleaner_apartments_apartment_id ON cleaner_apartments(apartment_id);
CREATE INDEX idx_cleaner_apartments_active ON cleaner_apartments(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_cleaner_apartments_assigned_at ON cleaner_apartments(assigned_at);

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER update_cleaners_updated_at 
    BEFORE UPDATE ON cleaners 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cleaner_apartments_updated_at 
    BEFORE UPDATE ON cleaner_apartments 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Комментарии к таблицам
COMMENT ON TABLE cleaners IS 'Уборщицы в системе';
COMMENT ON COLUMN cleaners.user_id IS 'Ссылка на пользователя-уборщицу';
COMMENT ON COLUMN cleaners.is_active IS 'Активна ли уборщица';
COMMENT ON COLUMN cleaners.schedule IS 'JSON с расписанием работы уборщицы';

COMMENT ON TABLE cleaner_apartments IS 'Связь между уборщицами и квартирами (many-to-many)';
COMMENT ON COLUMN cleaner_apartments.cleaner_id IS 'ID уборщицы';
COMMENT ON COLUMN cleaner_apartments.apartment_id IS 'ID квартиры';
COMMENT ON COLUMN cleaner_apartments.is_active IS 'Активна ли связь между уборщицей и квартирой';
COMMENT ON COLUMN cleaner_apartments.assigned_at IS 'Дата назначения уборщицы на квартиру';
