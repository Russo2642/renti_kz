-- Конвертация всех timestamp with time zone в timestamp (UTC)
-- Это решает проблему с часовыми поясами в scheduler

-- Обновляем таблицу bookings
ALTER TABLE bookings 
    ALTER COLUMN start_date TYPE TIMESTAMP USING start_date AT TIME ZONE 'UTC',
    ALTER COLUMN end_date TYPE TIMESTAMP USING end_date AT TIME ZONE 'UTC',
    ALTER COLUMN last_door_action TYPE TIMESTAMP USING last_door_action AT TIME ZONE 'UTC',
    ALTER COLUMN extension_end_date TYPE TIMESTAMP USING extension_end_date AT TIME ZONE 'UTC',
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

-- Обновляем таблицу booking_extensions
ALTER TABLE booking_extensions
    ALTER COLUMN requested_at TYPE TIMESTAMP USING requested_at AT TIME ZONE 'UTC',
    ALTER COLUMN approved_at TYPE TIMESTAMP USING approved_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMP USING updated_at AT TIME ZONE 'UTC';

-- Обновляем таблицу door_actions
ALTER TABLE door_actions
    ALTER COLUMN created_at TYPE TIMESTAMP USING created_at AT TIME ZONE 'UTC';

-- Обновляем функции, которые используют NOW() для работы с UTC
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = (NOW() AT TIME ZONE 'UTC');
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Изменяем DEFAULT значения для новых записей чтобы использовать UTC
ALTER TABLE bookings 
    ALTER COLUMN created_at SET DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ALTER COLUMN updated_at SET DEFAULT (NOW() AT TIME ZONE 'UTC');

ALTER TABLE booking_extensions
    ALTER COLUMN requested_at SET DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ALTER COLUMN created_at SET DEFAULT (NOW() AT TIME ZONE 'UTC'),
    ALTER COLUMN updated_at SET DEFAULT (NOW() AT TIME ZONE 'UTC');

ALTER TABLE door_actions
    ALTER COLUMN created_at SET DEFAULT (NOW() AT TIME ZONE 'UTC'); 