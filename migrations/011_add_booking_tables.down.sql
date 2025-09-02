-- Удаление триггеров
DROP TRIGGER IF EXISTS generate_booking_number_trigger ON bookings;
DROP TRIGGER IF EXISTS update_bookings_updated_at ON bookings;
DROP TRIGGER IF EXISTS update_booking_extensions_updated_at ON booking_extensions;

-- Удаление функций
DROP FUNCTION IF EXISTS generate_booking_number();
DROP FUNCTION IF EXISTS update_booking_number();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаление индексов
DROP INDEX IF EXISTS idx_bookings_renter_id;
DROP INDEX IF EXISTS idx_bookings_apartment_id;
DROP INDEX IF EXISTS idx_bookings_status;
DROP INDEX IF EXISTS idx_bookings_start_date;
DROP INDEX IF EXISTS idx_bookings_end_date;
DROP INDEX IF EXISTS idx_bookings_booking_number;
DROP INDEX IF EXISTS idx_bookings_dates_apartment;

DROP INDEX IF EXISTS idx_booking_extensions_booking_id;
DROP INDEX IF EXISTS idx_booking_extensions_status;

DROP INDEX IF EXISTS idx_door_actions_booking_id;
DROP INDEX IF EXISTS idx_door_actions_user_id;
DROP INDEX IF EXISTS idx_door_actions_created_at;

-- Удаление таблиц
DROP TABLE IF EXISTS door_actions;
DROP TABLE IF EXISTS booking_extensions;
DROP TABLE IF EXISTS bookings; 