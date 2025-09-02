-- Откат исправления функции daterange_overlaps

-- Удаляем исправленную функцию
DROP FUNCTION IF EXISTS daterange_overlaps(timestamp with time zone, timestamp with time zone, timestamp with time zone, timestamp with time zone);

-- Восстанавливаем старую функцию с неправильными типами
CREATE OR REPLACE FUNCTION daterange_overlaps(
    start1 timestamp, end1 timestamp,
    start2 timestamp, end2 timestamp
) RETURNS boolean AS $$
BEGIN
    RETURN (start1 < end2) AND (start2 < end1);
END;
$$ LANGUAGE plpgsql IMMUTABLE; 