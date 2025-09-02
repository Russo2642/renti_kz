-- Добавление новых типов уведомлений для операций с апартаментами
ALTER TYPE notification_type ADD VALUE 'apartment_created';
ALTER TYPE notification_type ADD VALUE 'apartment_approved';
ALTER TYPE notification_type ADD VALUE 'apartment_rejected';
ALTER TYPE notification_type ADD VALUE 'apartment_updated';
ALTER TYPE notification_type ADD VALUE 'apartment_status_changed';

-- Добавляем комментарии для документации новых типов
COMMENT ON TYPE notification_type IS 'Типы уведомлений: для бронирований (booking_*), для квартир (apartment_*), системные (password_ready, lock_issue и др.)'; 