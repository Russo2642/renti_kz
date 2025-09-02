-- Создание enum для статуса комнаты чата
CREATE TYPE chat_room_status AS ENUM (
    'pending',
    'active', 
    'closed',
    'archived'
);

-- Создание enum для типа сообщения
CREATE TYPE message_type AS ENUM (
    'text',
    'image',
    'file',
    'system',
    'welcome',
    'status'
);

-- Создание enum для статуса сообщения
CREATE TYPE message_status AS ENUM (
    'sent',
    'delivered',
    'read'
);

-- Создание enum для роли участника чата
CREATE TYPE chat_participant_role AS ENUM (
    'renter',
    'concierge',
    'moderator',
    'admin'
);

-- Таблица комнат чата
CREATE TABLE chat_rooms (
    id SERIAL PRIMARY KEY,
    booking_id INTEGER NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    concierge_id INTEGER NOT NULL REFERENCES concierges(id) ON DELETE CASCADE,
    renter_id INTEGER NOT NULL REFERENCES renters(id) ON DELETE CASCADE,
    apartment_id INTEGER NOT NULL REFERENCES apartments(id) ON DELETE CASCADE,
    status chat_room_status NOT NULL DEFAULT 'pending',
    opened_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    last_message_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ограничения
    CONSTRAINT unique_booking_chat UNIQUE(booking_id),
    CONSTRAINT check_dates CHECK (
        (opened_at IS NULL AND status = 'pending') OR 
        (opened_at IS NOT NULL AND status IN ('active', 'closed', 'archived'))
    )
);

-- Таблица сообщений в чате
CREATE TABLE chat_messages (
    id SERIAL PRIMARY KEY,
    chat_room_id INTEGER NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    sender_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type message_type NOT NULL DEFAULT 'text',
    content TEXT NOT NULL,
    file_url VARCHAR(500),
    file_name VARCHAR(255),
    file_size BIGINT,
    status message_status NOT NULL DEFAULT 'sent',
    read_at TIMESTAMPTZ,
    reply_to_id INTEGER REFERENCES chat_messages(id) ON DELETE SET NULL,
    edited_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ограничения
    CONSTRAINT check_file_fields CHECK (
        (type IN ('image', 'file') AND file_url IS NOT NULL) OR 
        (type NOT IN ('image', 'file') AND file_url IS NULL)
    ),
    CONSTRAINT check_content_not_empty CHECK (LENGTH(TRIM(content)) > 0)
);

-- Таблица участников чата
CREATE TABLE chat_participants (
    id SERIAL PRIMARY KEY,
    chat_room_id INTEGER NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role chat_participant_role NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at TIMESTAMPTZ,
    last_read_at TIMESTAMPTZ,
    is_online BOOLEAN NOT NULL DEFAULT FALSE,
    last_seen_at TIMESTAMPTZ,
    
    -- Ограничения
    CONSTRAINT unique_room_user UNIQUE(chat_room_id, user_id)
);

-- Индексы для оптимизации запросов
-- Chat rooms
CREATE INDEX idx_chat_rooms_booking_id ON chat_rooms(booking_id);
CREATE INDEX idx_chat_rooms_concierge_id ON chat_rooms(concierge_id);
CREATE INDEX idx_chat_rooms_renter_id ON chat_rooms(renter_id);
CREATE INDEX idx_chat_rooms_apartment_id ON chat_rooms(apartment_id);
CREATE INDEX idx_chat_rooms_status ON chat_rooms(status);
CREATE INDEX idx_chat_rooms_opened_at ON chat_rooms(opened_at) WHERE opened_at IS NOT NULL;
CREATE INDEX idx_chat_rooms_last_message_at ON chat_rooms(last_message_at) WHERE last_message_at IS NOT NULL;

-- Chat messages
CREATE INDEX idx_chat_messages_room_id ON chat_messages(chat_room_id);
CREATE INDEX idx_chat_messages_sender_id ON chat_messages(sender_id);
CREATE INDEX idx_chat_messages_created_at ON chat_messages(created_at);
CREATE INDEX idx_chat_messages_type ON chat_messages(type);
CREATE INDEX idx_chat_messages_status ON chat_messages(status);
CREATE INDEX idx_chat_messages_room_created ON chat_messages(chat_room_id, created_at);
CREATE INDEX idx_chat_messages_reply_to ON chat_messages(reply_to_id) WHERE reply_to_id IS NOT NULL;

-- Chat participants
CREATE INDEX idx_chat_participants_room_id ON chat_participants(chat_room_id);
CREATE INDEX idx_chat_participants_user_id ON chat_participants(user_id);
CREATE INDEX idx_chat_participants_online ON chat_participants(is_online) WHERE is_online = TRUE;
CREATE INDEX idx_chat_participants_role ON chat_participants(role);

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER update_chat_rooms_updated_at 
    BEFORE UPDATE ON chat_rooms 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chat_messages_updated_at 
    BEFORE UPDATE ON chat_messages 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Триггер для обновления last_message_at в комнате при новом сообщении
CREATE OR REPLACE FUNCTION update_chat_room_last_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE chat_rooms 
    SET last_message_at = NEW.created_at,
        updated_at = NOW()
    WHERE id = NEW.chat_room_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_chat_room_last_message
    AFTER INSERT ON chat_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_chat_room_last_message();

-- Комментарии к таблицам
COMMENT ON TABLE chat_rooms IS 'Комнаты чата между арендаторами и консьержами';
COMMENT ON COLUMN chat_rooms.booking_id IS 'Ссылка на бронирование';
COMMENT ON COLUMN chat_rooms.concierge_id IS 'Ссылка на консьержа';
COMMENT ON COLUMN chat_rooms.renter_id IS 'Ссылка на арендатора';
COMMENT ON COLUMN chat_rooms.status IS 'Статус комнаты чата';
COMMENT ON COLUMN chat_rooms.opened_at IS 'Время открытия чата (за 15 мин до брони)';
COMMENT ON COLUMN chat_rooms.closed_at IS 'Время закрытия чата (через 24 часа после брони)';

COMMENT ON TABLE chat_messages IS 'Сообщения в чате';
COMMENT ON COLUMN chat_messages.sender_id IS 'Отправитель сообщения';
COMMENT ON COLUMN chat_messages.type IS 'Тип сообщения';
COMMENT ON COLUMN chat_messages.reply_to_id IS 'Ссылка на сообщение, на которое отвечают';

COMMENT ON TABLE chat_participants IS 'Участники чата';
COMMENT ON COLUMN chat_participants.role IS 'Роль участника в чате'; 