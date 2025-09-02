-- Добавление роли консьержа в систему
INSERT INTO user_roles (name, description) VALUES
    ('concierge', 'Консьерж - обслуживает гостей и квартиры');

-- Комментарий к новой роли
COMMENT ON TABLE user_roles IS 'Роли пользователей в системе: admin, moderator, user, owner, concierge'; 