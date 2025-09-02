-- Откат удаления поля offered_price из таблицы bookings

-- Добавляем обратно поле offered_price
ALTER TABLE bookings ADD COLUMN offered_price INTEGER;

-- Добавляем комментарий к полю
COMMENT ON COLUMN bookings.offered_price IS 'Цена, предложенная пользователем при бронировании "по договорённости"'; 