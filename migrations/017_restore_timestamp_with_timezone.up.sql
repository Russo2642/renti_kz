-- Возвращаем TIMESTAMP WITH TIME ZONE для правильной работы с часовыми поясами
-- Это отменяет миграцию 015 и восстанавливает корректную работу с timezone

-- Изменяем колонки времени в таблице bookings
ALTER TABLE bookings 
    ALTER COLUMN start_date TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN end_date TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN last_door_action TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN extension_end_date TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE;

-- Изменяем колонки времени в таблице booking_extensions
ALTER TABLE booking_extensions 
    ALTER COLUMN requested_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN approved_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE;

-- Изменяем колонки времени в таблице door_actions
ALTER TABLE door_actions 
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE;

-- Восстанавливаем правильные DEFAULT значения
ALTER TABLE bookings 
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE booking_extensions
    ALTER COLUMN requested_at SET DEFAULT NOW(),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE door_actions
    ALTER COLUMN created_at SET DEFAULT NOW();

-- Восстанавливаем функцию триггера
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Обновляем комментарий
COMMENT ON TABLE bookings IS 'Времена хранятся в TIMESTAMP WITH TIME ZONE для корректной работы с часовыми поясами'; 