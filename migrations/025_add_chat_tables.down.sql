-- Удаление триггеров
DROP TRIGGER IF EXISTS trigger_update_chat_room_last_message ON chat_messages;
DROP TRIGGER IF EXISTS update_chat_messages_updated_at ON chat_messages;
DROP TRIGGER IF EXISTS update_chat_rooms_updated_at ON chat_rooms;

-- Удаление функции
DROP FUNCTION IF EXISTS update_chat_room_last_message();

-- Удаление таблиц (в обратном порядке из-за зависимостей)
DROP TABLE IF EXISTS chat_participants;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chat_rooms;

-- Удаление типов enum
DROP TYPE IF EXISTS chat_participant_role;
DROP TYPE IF EXISTS message_status;
DROP TYPE IF EXISTS message_type;
DROP TYPE IF EXISTS chat_room_status; 