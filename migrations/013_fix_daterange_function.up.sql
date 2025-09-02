-- Исправление функции daterange_overlaps для работы с timestamp with time zone

-- Удаляем старую функцию если существует
DROP FUNCTION IF EXISTS daterange_overlaps(timestamp, timestamp, timestamp, timestamp);

-- Создаем функцию с правильными типами
CREATE OR REPLACE FUNCTION daterange_overlaps(
    start1 timestamp with time zone, end1 timestamp with time zone,
    start2 timestamp with time zone, end2 timestamp with time zone
) RETURNS boolean AS $$
BEGIN
    RETURN (start1 < end2) AND (start2 < end1);
END;
$$ LANGUAGE plpgsql IMMUTABLE; 