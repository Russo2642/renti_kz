-- Откат ограничений на пересекающиеся бронирования

-- Удаляем триггер
DROP TRIGGER IF EXISTS booking_conflict_check ON bookings;

-- Удаляем функции
DROP FUNCTION IF EXISTS check_booking_conflicts();
DROP FUNCTION IF EXISTS daterange_overlaps(timestamp with time zone, timestamp with time zone, timestamp with time zone, timestamp with time zone); 