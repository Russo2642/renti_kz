-- Исправление часовых поясов в существующих бронированиях
-- Конвертируем времена из неправильно сохраненного UTC в правильное время
-- Предполагаем, что времена были сохранены как местное время Казахстана, но записаны как UTC

-- Исправляем времена в активных бронированиях (вычитаем 6 часов для Казахстана UTC+6)
UPDATE bookings 
SET 
    start_date = start_date - INTERVAL '5 hours',
    end_date = end_date - INTERVAL '5 hours',
    updated_at = NOW()
WHERE status IN ('pending', 'approved', 'active')
AND start_date > NOW() - INTERVAL '1 day'  -- Только недавние бронирования
AND EXTRACT(HOUR FROM start_date) BETWEEN 18 AND 23; -- Только те, что выглядят как местное время сохраненное как UTC

-- Добавляем комментарий о проделанной работе
COMMENT ON TABLE bookings IS 'Времена хранятся в UTC. Миграция 016 исправила часовые пояса в существующих данных'; 