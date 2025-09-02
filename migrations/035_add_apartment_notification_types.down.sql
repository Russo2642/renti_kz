-- Откат типов уведомлений для апартаментов
-- Примечание: PostgreSQL не поддерживает удаление значений из enum напрямую
-- Для полного отката необходимо пересоздать enum тип

-- 1. Создаем временный enum без значений для апартаментов  
CREATE TYPE notification_type_temp AS ENUM (
    'booking_approved',
    'booking_rejected', 
    'booking_canceled',
    'booking_completed',
    'password_ready',
    'extension_request',
    'extension_approved',
    'extension_rejected',
    'checkout_reminder',
    'lock_issue',
    'new_booking',
    'session_finished',
    'booking_starting_soon',
    'booking_ending',
    'payment_required'
);

-- 2. Обновляем колонку в таблице notifications
ALTER TABLE notifications ALTER COLUMN type TYPE notification_type_temp USING type::text::notification_type_temp;

-- 3. Удаляем старый enum
DROP TYPE notification_type;

-- 4. Переименовываем временный enum
ALTER TYPE notification_type_temp RENAME TO notification_type; 