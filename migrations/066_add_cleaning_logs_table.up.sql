-- Создание таблицы для логирования процесса уборки
CREATE TABLE cleaning_logs (
    id BIGSERIAL PRIMARY KEY,
    apartment_id INTEGER NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    cleaner_id INTEGER NOT NULL REFERENCES cleaners(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'completed', 'cancelled')),
    start_notes TEXT,
    completion_notes TEXT,
    photos_urls TEXT[], -- Массив URL фотографий после уборки
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_cleaning_logs_apartment_id ON cleaning_logs(apartment_id);
CREATE INDEX idx_cleaning_logs_cleaner_id ON cleaning_logs(cleaner_id);
CREATE INDEX idx_cleaning_logs_status ON cleaning_logs(status);
CREATE INDEX idx_cleaning_logs_started_at ON cleaning_logs(started_at);
CREATE INDEX idx_cleaning_logs_completed_at ON cleaning_logs(completed_at) WHERE completed_at IS NOT NULL;

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_cleaning_logs_updated_at 
    BEFORE UPDATE ON cleaning_logs 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Комментарии
COMMENT ON TABLE cleaning_logs IS 'Журнал процессов уборки квартир';
COMMENT ON COLUMN cleaning_logs.apartment_id IS 'ID квартиры';
COMMENT ON COLUMN cleaning_logs.cleaner_id IS 'ID уборщицы';
COMMENT ON COLUMN cleaning_logs.started_at IS 'Время начала уборки';
COMMENT ON COLUMN cleaning_logs.completed_at IS 'Время завершения уборки';
COMMENT ON COLUMN cleaning_logs.status IS 'Статус процесса уборки';
COMMENT ON COLUMN cleaning_logs.start_notes IS 'Заметки при начале уборки';
COMMENT ON COLUMN cleaning_logs.completion_notes IS 'Заметки при завершении уборки';
COMMENT ON COLUMN cleaning_logs.photos_urls IS 'URL фотографий после уборки';
