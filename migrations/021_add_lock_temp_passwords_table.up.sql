-- Добавляем поля для Tuya в таблицу замков
ALTER TABLE locks 
ADD COLUMN tuya_device_id VARCHAR(255),
ADD COLUMN owner_password VARCHAR(255);

-- Создаем таблицу для временных паролей замков
CREATE TABLE lock_temp_passwords (
    id SERIAL PRIMARY KEY,
    lock_id INTEGER NOT NULL REFERENCES locks(id) ON DELETE CASCADE,
    booking_id INTEGER REFERENCES bookings(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    password VARCHAR(255) NOT NULL,
    tuya_password_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создаем индексы для оптимизации запросов
CREATE INDEX idx_lock_temp_passwords_lock_id ON lock_temp_passwords(lock_id);
CREATE INDEX idx_lock_temp_passwords_booking_id ON lock_temp_passwords(booking_id);
CREATE INDEX idx_lock_temp_passwords_user_id ON lock_temp_passwords(user_id);
CREATE INDEX idx_lock_temp_passwords_valid_dates ON lock_temp_passwords(valid_from, valid_until);
CREATE INDEX idx_lock_temp_passwords_is_active ON lock_temp_passwords(is_active);

-- Добавляем триггер для обновления updated_at
CREATE OR REPLACE FUNCTION update_lock_temp_passwords_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_lock_temp_passwords_updated_at
    BEFORE UPDATE ON lock_temp_passwords
    FOR EACH ROW
    EXECUTE FUNCTION update_lock_temp_passwords_updated_at();

-- Создаем комментарии для таблицы и полей
COMMENT ON TABLE lock_temp_passwords IS 'Временные пароли для замков';
COMMENT ON COLUMN lock_temp_passwords.lock_id IS 'ID замка';
COMMENT ON COLUMN lock_temp_passwords.booking_id IS 'ID бронирования (если пароль для бронирования)';
COMMENT ON COLUMN lock_temp_passwords.user_id IS 'ID пользователя, для которого создан пароль';
COMMENT ON COLUMN lock_temp_passwords.password IS 'Временный пароль';
COMMENT ON COLUMN lock_temp_passwords.tuya_password_id IS 'ID пароля в системе Tuya';
COMMENT ON COLUMN lock_temp_passwords.name IS 'Название пароля';
COMMENT ON COLUMN lock_temp_passwords.valid_from IS 'Время начала действия пароля';
COMMENT ON COLUMN lock_temp_passwords.valid_until IS 'Время окончания действия пароля';
COMMENT ON COLUMN lock_temp_passwords.is_active IS 'Активен ли пароль';

-- Комментарии для новых полей в таблице замков
COMMENT ON COLUMN locks.tuya_device_id IS 'ID устройства в системе Tuya';
COMMENT ON COLUMN locks.owner_password IS 'Постоянный пароль владельца'; 