-- Добавление роли уборщицы в систему
INSERT INTO user_roles (name, description) VALUES
    ('cleaner', 'Уборщица/Уборщик - обслуживание и уборка квартир');

-- Комментарий к новой роли
COMMENT ON TABLE user_roles IS 'Роли пользователей в системе: admin, moderator, user, owner, concierge, cleaner';
