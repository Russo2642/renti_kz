-- Создание enum для типов уведомлений
CREATE TYPE notification_type AS ENUM (
    'booking_approved',
    'booking_rejected', 
    'booking_canceled',
    'booking_completed',
    'password_ready',
    'extension_request',
    'extension_approved',
    'extension_rejected',
    'checkout_reminder',
    'lock_issue',
    'new_booking',
    'session_finished',
    'booking_starting_soon',
    'booking_ending',
    'payment_required'
);

-- Создание enum для приоритетов уведомлений
CREATE TYPE notification_priority AS ENUM (
    'low',
    'normal',
    'high',
    'urgent'
);

-- Создание enum для типов устройств
CREATE TYPE device_type AS ENUM (
    'ios',
    'android',
    'web'
);

-- Таблица уведомлений
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    priority notification_priority NOT NULL DEFAULT 'normal',
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    is_pushed BOOLEAN NOT NULL DEFAULT FALSE,
    booking_id INTEGER REFERENCES bookings(id) ON DELETE SET NULL,
    apartment_id INTEGER REFERENCES apartments(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read_at TIMESTAMPTZ
);

-- Таблица устройств пользователей для push-уведомлений
CREATE TABLE user_devices (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token VARCHAR(255) NOT NULL UNIQUE,
    device_type device_type NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    app_version VARCHAR(50),
    os_version VARCHAR(50),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_user_unread ON notifications(user_id, is_read) WHERE is_read = FALSE;
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_booking_id ON notifications(booking_id) WHERE booking_id IS NOT NULL;
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

CREATE INDEX idx_user_devices_user_id ON user_devices(user_id);
CREATE INDEX idx_user_devices_token ON user_devices(device_token);
CREATE INDEX idx_user_devices_active ON user_devices(is_active) WHERE is_active = TRUE;

-- Триггер для автоматического обновления updated_at в user_devices
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_devices_updated_at 
    BEFORE UPDATE ON user_devices 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 