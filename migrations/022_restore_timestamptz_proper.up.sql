-- Возвращаем TIMESTAMP WITH TIME ZONE для правильной работы с часовыми поясами
-- Этого требует правильная архитектура: БД хранит всё в UTC, приложение работает с UTC+5

-- Обновляем таблицу bookings - возвращаем TIMESTAMP WITH TIME ZONE
ALTER TABLE bookings 
    ALTER COLUMN start_date TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN end_date TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN last_door_action TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN extension_end_date TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE;

-- Обновляем таблицу booking_extensions
ALTER TABLE booking_extensions 
    ALTER COLUMN requested_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN approved_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE,
    ALTER COLUMN updated_at TYPE TIMESTAMP WITH TIME ZONE;

-- Обновляем таблицу door_actions
ALTER TABLE door_actions 
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE;

-- Обновляем DEFAULT значения - используем NOW() для автоматического UTC
ALTER TABLE bookings 
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE booking_extensions
    ALTER COLUMN requested_at SET DEFAULT NOW(),
    ALTER COLUMN created_at SET DEFAULT NOW(),
    ALTER COLUMN updated_at SET DEFAULT NOW();

ALTER TABLE door_actions
    ALTER COLUMN created_at SET DEFAULT NOW();

-- Обновляем функцию триггера для корректной работы с UTC
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Обновляем комментарий согласно новой архитектуре
COMMENT ON TABLE bookings IS 'Архитектура времени: БД хранит всё в UTC (TIMESTAMPTZ), ввод от пользователя (UTC+5) конвертируется в UTC, вывод пользователю из UTC в UTC+5, планировщик работает с UTC';

-- Устанавливаем сессионную временную зону в UTC для PostgreSQL
-- Это гарантирует что все NOW() будут в UTC
-- Используем имя базы данных из конфигурации
ALTER DATABASE renti_kz SET timezone TO 'UTC'; 