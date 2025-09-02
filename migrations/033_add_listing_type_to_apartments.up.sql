-- Добавление поля listing_type в таблицу apartments
-- Это поле будет указывать, кто размещает объявление: 'owner' (от хозяина) или 'realtor' (от риэлтора)
ALTER TABLE apartments ADD COLUMN listing_type VARCHAR(10) NOT NULL DEFAULT 'owner';

-- Добавление проверки для корректных значений
ALTER TABLE apartments ADD CONSTRAINT chk_listing_type CHECK (listing_type IN ('owner', 'realtor'));

-- Создание индекса для лучшей производительности при фильтрации
CREATE INDEX idx_apartments_listing_type ON apartments(listing_type);

-- Комментарий для поля
COMMENT ON COLUMN apartments.listing_type IS 'Тип размещения объявления: owner (от хозяина) или realtor (от риэлтора)'; 