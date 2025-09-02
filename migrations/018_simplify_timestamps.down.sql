-- Откат упрощения - возвращаем TIMESTAMP WITH TIME ZONE

-- Возвращаем колонки времени в таблице bookings к TIMESTAMP WITH TIME ZONE
ALTER TABLE bookings 
    ALTER COLUMN start_date TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN end_date TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN last_door_action TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN extension_end_date TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE;

-- Возвращаем колонки времени в таблице booking_extensions
ALTER TABLE booking_extensions 
    ALTER COLUMN requested_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN approved_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE;

-- Возвращаем колонки времени в таблице door_actions
ALTER TABLE door_actions 
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE;

-- Возвращаем старый комментарий
COMMENT ON TABLE bookings IS 'Времена хранятся в TIMESTAMP WITH TIME ZONE для корректной работы с часовыми поясами'; 