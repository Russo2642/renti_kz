-- Откат миграции 022 - возвращаем простой TIMESTAMP

-- Возвращаем таблицу bookings к простому TIMESTAMP
ALTER TABLE bookings 
    ALTER COLUMN start_date TYPE TIMESTAMP,
    ALTER COLUMN end_date TYPE TIMESTAMP,
    ALTER COLUMN last_door_action TYPE TIMESTAMP,
    ALTER COLUMN extension_end_date TYPE TIMESTAMP,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Возвращаем таблицу booking_extensions к простому TIMESTAMP
ALTER TABLE booking_extensions 
    ALTER COLUMN requested_at TYPE TIMESTAMP,
    ALTER COLUMN approved_at TYPE TIMESTAMP,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Возвращаем таблицу door_actions к простому TIMESTAMP
ALTER TABLE door_actions 
    ALTER COLUMN created_at TYPE TIMESTAMP;

-- Возвращаем старую функцию триггера
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Убираем комментарий
COMMENT ON TABLE bookings IS 'Времена хранятся в простом TIMESTAMP, планировщик работает с локальным временем сервера';

-- Возвращаем часовой пояс по умолчанию
ALTER DATABASE renti_kz RESET timezone; 