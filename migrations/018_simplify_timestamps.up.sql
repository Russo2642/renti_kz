-- Упрощение работы с временем
-- Убираем TIMESTAMP WITH TIME ZONE и делаем просто TIMESTAMP
-- Планировщик будет работать с локальным временем сервера

-- Изменяем колонки времени в таблице bookings на простой TIMESTAMP
ALTER TABLE bookings 
    ALTER COLUMN start_date TYPE TIMESTAMP,
    ALTER COLUMN end_date TYPE TIMESTAMP,
    ALTER COLUMN last_door_action TYPE TIMESTAMP,
    ALTER COLUMN extension_end_date TYPE TIMESTAMP,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Изменяем колонки времени в таблице booking_extensions
ALTER TABLE booking_extensions 
    ALTER COLUMN requested_at TYPE TIMESTAMP,
    ALTER COLUMN approved_at TYPE TIMESTAMP,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Изменяем колонки времени в таблице door_actions
ALTER TABLE door_actions 
    ALTER COLUMN created_at TYPE TIMESTAMP;

-- Упрощаем DEFAULT значения - используем простой NOW()
ALTER TABLE bookings 
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE booking_extensions
    ALTER COLUMN requested_at SET DEFAULT NOW(),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE door_actions
    ALTER COLUMN created_at SET DEFAULT NOW();

-- Упрощаем функцию триггера
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Обновляем комментарий
COMMENT ON TABLE bookings IS 'Времена хранятся в простом TIMESTAMP, планировщик работает с локальным временем сервера'; 