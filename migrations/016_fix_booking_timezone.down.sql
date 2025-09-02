-- Откат исправления часовых поясов
-- Возвращаем времена обратно (добавляем 5 часов)

UPDATE bookings 
SET 
    start_date = start_date + INTERVAL '5 hours',
    end_date = end_date + INTERVAL '5 hours',
    updated_at = NOW()
WHERE status IN ('pending', 'approved', 'active')
AND start_date > NOW() - INTERVAL '2 days'
AND EXTRACT(HOUR FROM start_date) BETWEEN 12 AND 17; -- Времена после исправления

-- Убираем комментарий
COMMENT ON TABLE bookings IS NULL; 