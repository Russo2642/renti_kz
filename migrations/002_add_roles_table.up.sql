-- Создание таблицы ролей пользователей
CREATE TABLE IF NOT EXISTS user_roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) UNIQUE NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Вставка предопределенных ролей
INSERT INTO user_roles (name, description) VALUES
    ('admin', 'Администратор системы'),
    ('moderator', 'Модератор системы'),
    ('user', 'Обычный пользователь (арендатор)'),
    ('owner', 'Владелец недвижимости');

-- Добавление временного столбца для хранения id роли
ALTER TABLE users ADD COLUMN role_id INT;

-- Обновление временного столбца на основании существующих данных
UPDATE users SET role_id = (SELECT id FROM user_roles WHERE name = 
    CASE 
        WHEN role = 'client' THEN 'user' 
        ELSE role 
    END
);

-- Создание внешнего ключа и ограничения NOT NULL
ALTER TABLE users 
    ALTER COLUMN role_id SET NOT NULL,
    ADD CONSTRAINT fk_user_role FOREIGN KEY (role_id) REFERENCES user_roles(id);

-- Удаление старого столбца role
ALTER TABLE users DROP COLUMN role;

-- Обновление индекса
DROP INDEX IF EXISTS idx_users_role;
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id); 