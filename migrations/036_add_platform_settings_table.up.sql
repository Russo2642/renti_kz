-- Создание таблицы для настроек платформы
CREATE TABLE platform_settings (
    id SERIAL PRIMARY KEY,
    setting_key VARCHAR(255) NOT NULL UNIQUE,
    setting_value VARCHAR(1000) NOT NULL,
    description TEXT,
    data_type VARCHAR(50) NOT NULL DEFAULT 'string', -- string, integer, decimal, boolean
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Создание индекса для быстрого поиска по ключу
CREATE INDEX idx_platform_settings_key ON platform_settings(setting_key);
CREATE INDEX idx_platform_settings_active ON platform_settings(is_active);

-- Добавление базовых настроек
INSERT INTO platform_settings (setting_key, setting_value, description, data_type, is_active) VALUES
('service_fee_percentage', '15', 'Процент сервисного сбора платформы', 'integer', true),
('min_booking_duration_hours', '1', 'Минимальная продолжительность бронирования в часах', 'integer', true),
('max_booking_duration_hours', '720', 'Максимальная продолжительность бронирования в часах (30 дней)', 'integer', true),
('default_cleaning_duration_minutes', '60', 'Стандартное время уборки между бронированиями в минутах', 'integer', true),
('platform_commission_percentage', '5', 'Комиссия платформы с владельцев недвижимости', 'integer', true),
('max_advance_booking_days', '90', 'Максимальное количество дней для предварительного бронирования', 'integer', true);

-- Добавление комментария к таблице
COMMENT ON TABLE platform_settings IS 'Настройки платформы для управления поведением системы'; 