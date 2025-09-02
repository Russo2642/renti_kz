-- Создание таблицы бронирований
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    renter_id INTEGER NOT NULL REFERENCES renters(id) ON DELETE CASCADE,
    apartment_id INTEGER NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    duration INTEGER NOT NULL, -- в часах
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'active', 'completed', 'canceled')),
    total_price INTEGER NOT NULL DEFAULT 0, -- базовая стоимость аренды
    service_fee INTEGER NOT NULL DEFAULT 0, -- сервисный сбор
    final_price INTEGER NOT NULL DEFAULT 0, -- итоговая стоимость
    is_contract_accepted BOOLEAN NOT NULL DEFAULT false, -- согласие на договор аренды
    cancellation_reason TEXT, -- причина отмены
    owner_comment TEXT, -- комментарий владельца при отклонении
    booking_number VARCHAR(50) UNIQUE NOT NULL, -- уникальный номер бронирования
    door_status VARCHAR(10) NOT NULL DEFAULT 'closed' CHECK (door_status IN ('open', 'closed')),
    last_door_action TIMESTAMP WITH TIME ZONE, -- время последнего действия с замком
    can_extend BOOLEAN NOT NULL DEFAULT true, -- можно ли продлить аренду
    extension_requested BOOLEAN NOT NULL DEFAULT false, -- запрошено ли продление
    extension_end_date TIMESTAMP WITH TIME ZONE, -- новая дата окончания при продлении
    extension_duration INTEGER DEFAULT 0, -- дополнительная продолжительность в часах
    extension_price INTEGER DEFAULT 0, -- стоимость продления
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создание таблицы продлений аренды
CREATE TABLE booking_extensions (
    id SERIAL PRIMARY KEY,
    booking_id INTEGER NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    duration INTEGER NOT NULL, -- дополнительная продолжительность в часах
    price INTEGER NOT NULL, -- стоимость продления
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    requested_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMP WITH TIME ZONE, -- время подтверждения
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создание таблицы действий с электронным замком
CREATE TABLE door_actions (
    id SERIAL PRIMARY KEY,
    booking_id INTEGER NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(10) NOT NULL CHECK (action IN ('open', 'closed')),
    success BOOLEAN NOT NULL DEFAULT false,
    error TEXT, -- описание ошибки если есть
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создание индексов для оптимизации запросов
CREATE INDEX idx_bookings_renter_id ON bookings(renter_id);
CREATE INDEX idx_bookings_apartment_id ON bookings(apartment_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_start_date ON bookings(start_date);
CREATE INDEX idx_bookings_end_date ON bookings(end_date);
CREATE INDEX idx_bookings_booking_number ON bookings(booking_number);
CREATE INDEX idx_bookings_dates_apartment ON bookings(apartment_id, start_date, end_date);

CREATE INDEX idx_booking_extensions_booking_id ON booking_extensions(booking_id);
CREATE INDEX idx_booking_extensions_status ON booking_extensions(status);

CREATE INDEX idx_door_actions_booking_id ON door_actions(booking_id);
CREATE INDEX idx_door_actions_user_id ON door_actions(user_id);
CREATE INDEX idx_door_actions_created_at ON door_actions(created_at);

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_bookings_updated_at BEFORE UPDATE ON bookings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_booking_extensions_updated_at BEFORE UPDATE ON booking_extensions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Функция для генерации уникального номера бронирования
CREATE OR REPLACE FUNCTION generate_booking_number()
RETURNS TRIGGER AS $$
BEGIN
    NEW.booking_number = 'AD' || LPAD(NEW.id::TEXT, 3, '0');
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Создание триггера AFTER INSERT для генерации номера бронирования
CREATE TRIGGER generate_booking_number_trigger 
    AFTER INSERT ON bookings
    FOR EACH ROW 
    EXECUTE FUNCTION generate_booking_number();

-- Обновление booking_number после вставки
CREATE OR REPLACE FUNCTION update_booking_number()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE bookings SET booking_number = 'AD' || LPAD(NEW.id::TEXT, 3, '0') WHERE id = NEW.id;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS generate_booking_number_trigger ON bookings;
CREATE TRIGGER generate_booking_number_trigger 
    AFTER INSERT ON bookings
    FOR EACH ROW 
    EXECUTE FUNCTION update_booking_number(); 