-- Добавление ограничений для предотвращения пересекающихся активных бронирований

-- Создаем функцию для проверки пересечения периодов
CREATE OR REPLACE FUNCTION daterange_overlaps(
    start1 timestamp with time zone, end1 timestamp with time zone,
    start2 timestamp with time zone, end2 timestamp with time zone
) RETURNS boolean AS $$
BEGIN
    RETURN (start1 < end2) AND (start2 < end1);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Создаем триггерную функцию для проверки конфликтов бронирований
CREATE OR REPLACE FUNCTION check_booking_conflicts() RETURNS trigger AS $$
BEGIN
    -- Проверяем конфликты только для активных статусов бронирований
    IF NEW.status IN ('pending', 'approved', 'active') THEN
        -- Проверяем наличие пересекающихся бронирований
        IF EXISTS (
            SELECT 1 FROM bookings 
            WHERE apartment_id = NEW.apartment_id 
            AND id != COALESCE(NEW.id, 0)
            AND status IN ('pending', 'approved', 'active')
            AND daterange_overlaps(start_date, end_date, NEW.start_date, NEW.end_date)
        ) THEN
            RAISE EXCEPTION 'Apartment is not available for the selected period. Conflicting booking exists.';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Создаем триггер для проверки при INSERT и UPDATE
DROP TRIGGER IF EXISTS booking_conflict_check ON bookings;
CREATE TRIGGER booking_conflict_check
    BEFORE INSERT OR UPDATE ON bookings
    FOR EACH ROW
    EXECUTE FUNCTION check_booking_conflicts();

-- Добавляем комментарий к таблице
COMMENT ON TRIGGER booking_conflict_check ON bookings IS 'Prevents overlapping active bookings for the same apartment'; 