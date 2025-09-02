-- Удаление роли уборщицы
DELETE FROM user_roles WHERE name = 'cleaner';

-- Восстановление комментария
COMMENT ON TABLE user_roles IS 'Роли пользователей в системе: admin, moderator, user, owner, concierge';
