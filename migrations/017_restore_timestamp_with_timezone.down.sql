-- Откат миграции 017 - возвращаем TIMESTAMP без timezone

-- Изменяем колонки времени в таблице bookings обратно
ALTER TABLE bookings 
    ALTER COLUMN start_date TYPE TIMESTAMP,
    ALTER COLUMN end_date TYPE TIMESTAMP,
    ALTER COLUMN last_door_action TYPE TIMESTAMP,
    ALTER COLUMN extension_end_date TYPE TIMESTAMP,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Изменяем колонки времени в таблице booking_extensions обратно
ALTER TABLE booking_extensions 
    ALTER COLUMN requested_at TYPE TIMESTAMP,
    ALTER COLUMN approved_at TYPE TIMESTAMP,
    ALTER COLUMN created_at TYPE TIMESTAMP,
    ALTER COLUMN updated_at TYPE TIMESTAMP;

-- Изменяем колонки времени в таблице door_actions обратно
ALTER TABLE door_actions 
    ALTER COLUMN created_at TYPE TIMESTAMP;

-- Возвращаем UTC DEFAULT значения
ALTER TABLE bookings 
    ALTER COLUMN created_at SET DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ALTER COLUMN updated_at SET DEFAULT (NOW() AT TIME ZONE 'UTC');

ALTER TABLE booking_extensions
    ALTER COLUMN requested_at SET DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ALTER COLUMN created_at SET DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ALTER COLUMN updated_at SET DEFAULT (NOW() AT TIME ZONE 'UTC');

ALTER TABLE door_actions
    ALTER COLUMN created_at SET DEFAULT (NOW() AT TIME ZONE 'UTC');

-- Возвращаем UTC функцию триггера
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = (NOW() AT TIME ZONE 'UTC');
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Возвращаем старый комментарий
COMMENT ON TABLE bookings IS 'Времена хранятся в UTC. Миграция 016 исправила часовые пояса в существующих данных'; 