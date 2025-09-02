-- Добавление поля is_active в таблицу users
-- Пользователи с is_active = false могут только просматривать публичные данные

BEGIN;

-- Добавляем поле is_active со значением по умолчанию true
ALTER TABLE users 
ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- Создаем индекс для быстрого поиска активных пользователей
CREATE INDEX idx_users_is_active ON users(is_active);

-- Добавляем комментарий к полю
COMMENT ON COLUMN users.is_active IS 'Статус активности пользователя. false = заблокированный пользователь с ограниченными правами';

COMMIT; 