-- Добавление столбца role
ALTER TABLE users ADD COLUMN role VARCHAR(20);

-- Обновление столбца role на основании role_id
UPDATE users SET role = (
    SELECT name FROM user_roles WHERE id = role_id
);

-- Заменяем 'user' на 'client'
UPDATE users SET role = 'client' WHERE role = 'user';

-- Установка ограничения NOT NULL
ALTER TABLE users ALTER COLUMN role SET NOT NULL;

-- Удаление внешнего ключа и столбца role_id
ALTER TABLE users DROP CONSTRAINT fk_user_role;
ALTER TABLE users DROP COLUMN role_id;

-- Восстановление индекса
DROP INDEX IF EXISTS idx_users_role_id;
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Удаление таблицы ролей
DROP TABLE IF EXISTS user_roles; 