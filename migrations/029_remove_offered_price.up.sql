-- Удаление поля offered_price из таблицы bookings
-- Это поле использовалось для бронирований "по договорённости", которые больше не поддерживаются

-- Удаляем поле offered_price
ALTER TABLE bookings DROP COLUMN IF EXISTS offered_price; 