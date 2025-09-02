-- Добавление поля цены за сутки в таблицу apartments
ALTER TABLE apartments ADD COLUMN daily_price INTEGER NOT NULL DEFAULT 0;

-- Добавление комментария к новой колонке
COMMENT ON COLUMN apartments.daily_price IS 'Цена за сутки (24 часа) аренды квартиры в тенге';

-- Добавление ограничения: daily_price не может быть отрицательным
ALTER TABLE apartments ADD CONSTRAINT check_daily_price_non_negative 
CHECK (daily_price >= 0);

-- Добавление индекса для эффективного поиска по цене за сутки
CREATE INDEX idx_apartments_daily_price ON apartments(daily_price); 