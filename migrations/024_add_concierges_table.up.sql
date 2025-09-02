-- Таблица консьержей
CREATE TABLE concierges (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    apartment_id INTEGER NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    schedule JSONB, -- расписание работы консьержа
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ограничения
    CONSTRAINT unique_user_concierge UNIQUE(user_id),
    CONSTRAINT unique_apartment_concierge UNIQUE(apartment_id)
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_concierges_user_id ON concierges(user_id);
CREATE INDEX idx_concierges_apartment_id ON concierges(apartment_id);
CREATE INDEX idx_concierges_active ON concierges(is_active) WHERE is_active = TRUE;

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_concierges_updated_at 
    BEFORE UPDATE ON concierges 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Комментарии к таблице
COMMENT ON TABLE concierges IS 'Консьержи, прикрепленные к квартирам';
COMMENT ON COLUMN concierges.user_id IS 'Ссылка на пользователя-консьержа';
COMMENT ON COLUMN concierges.apartment_id IS 'Ссылка на квартиру';
COMMENT ON COLUMN concierges.schedule IS 'JSON с расписанием работы консьержа';
COMMENT ON COLUMN concierges.is_active IS 'Активен ли консьерж'; 