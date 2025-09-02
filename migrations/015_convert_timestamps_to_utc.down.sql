-- Откат: возвращаем timestamp with time zone

-- Возвращаем DEFAULT значения
ALTER TABLE door_actions
    ALTER COLUMN created_at SET DEFAULT NOW();

ALTER TABLE booking_extensions
    ALTER COLUMN requested_at SET DEFAULT NOW(),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE bookings 
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

-- Возвращаем функцию update_updated_at_column к исходному виду
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Возвращаем типы колонок к timestamp with time zone
ALTER TABLE door_actions
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE USING created_at AT TIME ZONE 'UTC';

ALTER TABLE booking_extensions
    ALTER COLUMN requested_at TYPE TIMESTAMP WITH TIME ZONE USING requested_at AT TIME ZONE 'UTC',
    ALTER COLUMN approved_at TYPE TIMESTAMP WITH TIME ZONE USING approved_at AT TIME ZONE 'UTC',
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE USING updated_at AT TIME ZONE 'UTC';

ALTER TABLE bookings 
    ALTER COLUMN start_date TYPE TIMESTAMP WITH TIME ZONE USING start_date AT TIME ZONE 'UTC',
    ALTER COLUMN end_date TYPE TIMESTAMP WITH TIME ZONE USING end_date AT TIME ZONE 'UTC',
    ALTER COLUMN last_door_action TYPE TIMESTAMP WITH TIME ZONE USING last_door_action AT TIME ZONE 'UTC',
    ALTER COLUMN extension_end_date TYPE TIMESTAMP WITH TIME ZONE USING extension_end_date AT TIME ZONE 'UTC',
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE USING updated_at AT TIME ZONE 'UTC'; 