-- Добавление типов аренды в таблицу apartments
ALTER TABLE apartments ADD COLUMN rental_type_hourly BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE apartments ADD COLUMN rental_type_daily BOOLEAN NOT NULL DEFAULT true;

-- Добавление комментариев к новым колонкам
COMMENT ON COLUMN apartments.rental_type_hourly IS 'Поддерживает ли квартира почасовую аренду';
COMMENT ON COLUMN apartments.rental_type_daily IS 'Поддерживает ли квартира посуточную аренду';

-- Добавление ограничения: хотя бы один тип аренды должен быть выбран
ALTER TABLE apartments ADD CONSTRAINT check_rental_type_selected 
CHECK (rental_type_hourly = true OR rental_type_daily = true);

-- Добавление индексов для эффективного поиска по типам аренды
CREATE INDEX idx_apartments_rental_type_hourly ON apartments(rental_type_hourly);
CREATE INDEX idx_apartments_rental_type_daily ON apartments(rental_type_daily);
CREATE INDEX idx_apartments_rental_types ON apartments(rental_type_hourly, rental_type_daily); 