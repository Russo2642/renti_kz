-- Создание таблицы замков
CREATE TABLE locks (
    id SERIAL PRIMARY KEY,
    unique_id VARCHAR(50) UNIQUE NOT NULL, -- Уникальный ID замка (например, LOCK001, LOCK002)
    apartment_id INTEGER NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL, -- Название замка (например, "Замок квартиры 12")
    description TEXT, -- Описание замка
    current_status VARCHAR(10) NOT NULL DEFAULT 'closed' CHECK (current_status IN ('open', 'closed')),
    last_status_update TIMESTAMP WITH TIME ZONE, -- Время последнего обновления статуса
    last_heartbeat TIMESTAMP WITH TIME ZONE, -- Время последнего "сердцебиения"
    is_online BOOLEAN NOT NULL DEFAULT false, -- Онлайн ли замок
    firmware_version VARCHAR(50), -- Версия прошивки
    battery_level INTEGER CHECK (battery_level >= 0 AND battery_level <= 100), -- Уровень заряда батареи (0-100%)
    signal_strength INTEGER CHECK (signal_strength >= -100 AND signal_strength <= 0), -- Сила сигнала WiFi/GSM в dBm
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создание таблицы логов статуса замков (для истории изменений)
CREATE TABLE lock_status_logs (
    id SERIAL PRIMARY KEY,
    lock_id INTEGER NOT NULL REFERENCES locks(id) ON DELETE CASCADE,
    old_status VARCHAR(10) CHECK (old_status IN ('open', 'closed')),
    new_status VARCHAR(10) NOT NULL CHECK (new_status IN ('open', 'closed')),
    change_source VARCHAR(20) NOT NULL CHECK (change_source IN ('api', 'manual', 'system')), -- Источник изменения
    user_id INTEGER REFERENCES users(id), -- ID пользователя (если изменение через API)
    booking_id INTEGER REFERENCES bookings(id), -- ID бронирования (если через API)
    notes TEXT, -- Дополнительные заметки
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создание индексов для оптимизации
CREATE INDEX idx_locks_apartment_id ON locks(apartment_id);
CREATE INDEX idx_locks_unique_id ON locks(unique_id);
CREATE INDEX idx_locks_status ON locks(current_status);
CREATE INDEX idx_locks_online ON locks(is_online);
CREATE INDEX idx_locks_heartbeat ON locks(last_heartbeat);

CREATE INDEX idx_lock_status_logs_lock_id ON lock_status_logs(lock_id);
CREATE INDEX idx_lock_status_logs_created_at ON lock_status_logs(created_at);
CREATE INDEX idx_lock_status_logs_source ON lock_status_logs(change_source);

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_locks_updated_at BEFORE UPDATE ON locks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Триггер для автоматического создания логов при изменении статуса
CREATE OR REPLACE FUNCTION log_lock_status_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Записываем лог только если статус действительно изменился
    IF NEW.current_status != OLD.current_status THEN
        INSERT INTO lock_status_logs (lock_id, old_status, new_status, change_source, notes)
        VALUES (NEW.id, OLD.current_status, NEW.current_status, 'system', 'Автоматическое логирование изменения статуса');
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER log_lock_status_change_trigger 
    AFTER UPDATE ON locks
    FOR EACH ROW 
    EXECUTE FUNCTION log_lock_status_change();

-- Добавляем поле lock_id в таблицу apartments (опционально, для быстрого доступа)
ALTER TABLE apartments ADD COLUMN lock_id INTEGER REFERENCES locks(id);
CREATE INDEX idx_apartments_lock_id ON apartments(lock_id); 