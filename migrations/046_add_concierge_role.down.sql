-- Удаление роли консьержа из системы
-- Сначала обновляем всех пользователей с ролью concierge на user
UPDATE users SET role_id = (SELECT id FROM user_roles WHERE name = 'user') 
WHERE role_id = (SELECT id FROM user_roles WHERE name = 'concierge');

-- Удаляем роль concierge
DELETE FROM user_roles WHERE name = 'concierge';

-- Возвращаем комментарий
COMMENT ON TABLE user_roles IS 'Роли пользователей в системе: admin, moderator, user, owner'; 