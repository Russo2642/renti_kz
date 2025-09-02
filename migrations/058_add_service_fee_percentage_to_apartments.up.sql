-- Добавление поля процента сервисного сбора в таблицу apartments
ALTER TABLE apartments ADD COLUMN service_fee_percentage INTEGER NOT NULL DEFAULT 0;

-- Добавление комментария к новой колонке
COMMENT ON COLUMN apartments.service_fee_percentage IS 'Процент сервисного сбора для квартиры (может быть индивидуальным)';

-- Добавление ограничения: service_fee_percentage должен быть от 0 до 100
ALTER TABLE apartments ADD CONSTRAINT check_service_fee_percentage_range 
CHECK (service_fee_percentage >= 0 AND service_fee_percentage <= 100);

-- Добавление индекса для эффективного поиска по проценту сервисного сбора
CREATE INDEX idx_apartments_service_fee_percentage ON apartments(service_fee_percentage);

-- Обновляем все существующие квартиры значением по умолчанию из platform_settings
-- Берем значение из настроек платформы
UPDATE apartments 
SET service_fee_percentage = (
    SELECT COALESCE(CAST(setting_value AS INTEGER), 15) 
    FROM platform_settings 
    WHERE setting_key = 'service_fee_percentage' 
    AND is_active = true
    LIMIT 1
)
WHERE service_fee_percentage = 0;