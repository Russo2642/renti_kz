-- Рефакторинг архитектуры консьержей для поддержки нескольких квартир

-- Удаляем ограничение на apartment_id (если существует)
ALTER TABLE concierges DROP COLUMN IF EXISTS apartment_id CASCADE;

-- Создаем промежуточную таблицу для связи many-to-many между консьержами и квартирами
CREATE TABLE concierge_apartments (
    id SERIAL PRIMARY KEY,
    concierge_id INTEGER NOT NULL REFERENCES concierges(id) ON DELETE CASCADE,
    apartment_id INTEGER NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Уникальное ограничение: один консьерж может быть привязан к одной квартире только один раз активно
    CONSTRAINT unique_active_concierge_apartment UNIQUE (concierge_id, apartment_id, is_active) DEFERRABLE INITIALLY DEFERRED
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_concierge_apartments_concierge_id ON concierge_apartments(concierge_id);
CREATE INDEX idx_concierge_apartments_apartment_id ON concierge_apartments(apartment_id);
CREATE INDEX idx_concierge_apartments_active ON concierge_apartments(is_active);
CREATE INDEX idx_concierge_apartments_assigned_at ON concierge_apartments(assigned_at);

-- Комментарии для документации
COMMENT ON TABLE concierge_apartments IS 'Связь между консьержами и квартирами (many-to-many)';
COMMENT ON COLUMN concierge_apartments.concierge_id IS 'ID консьержа';
COMMENT ON COLUMN concierge_apartments.apartment_id IS 'ID квартиры';
COMMENT ON COLUMN concierge_apartments.is_active IS 'Активна ли связь между консьержем и квартирой';
COMMENT ON COLUMN concierge_apartments.assigned_at IS 'Дата назначения консьержа на квартиру';

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_concierge_apartments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER trigger_update_concierge_apartments_updated_at
    BEFORE UPDATE ON concierge_apartments
    FOR EACH ROW
    EXECUTE FUNCTION update_concierge_apartments_updated_at();