-- Создание таблицы для хранения документов квартиры
CREATE TABLE IF NOT EXISTS apartment_documents (
    id SERIAL PRIMARY KEY,
    apartment_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_apartment_documents_apartment FOREIGN KEY (apartment_id) REFERENCES apartments (id) ON DELETE CASCADE
);

-- Создание индекса для быстрого поиска документов по ID квартиры
CREATE INDEX IF NOT EXISTS idx_apartment_documents_apartment_id ON apartment_documents (apartment_id);

-- Добавление триггера для автоматического обновления поля updated_at
CREATE OR REPLACE FUNCTION update_apartment_documents_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_apartment_documents_updated_at
BEFORE UPDATE ON apartment_documents
FOR EACH ROW
EXECUTE FUNCTION update_apartment_documents_updated_at();

COMMENT ON TABLE apartment_documents IS 'Документы квартиры';
COMMENT ON COLUMN apartment_documents.id IS 'Уникальный идентификатор';
COMMENT ON COLUMN apartment_documents.apartment_id IS 'ID квартиры';
COMMENT ON COLUMN apartment_documents.url IS 'URL документа в хранилище';
COMMENT ON COLUMN apartment_documents.type IS 'Тип документа';
COMMENT ON COLUMN apartment_documents.created_at IS 'Время создания записи';
COMMENT ON COLUMN apartment_documents.updated_at IS 'Время последнего обновления записи'; 