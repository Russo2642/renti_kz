-- Создание таблицы правил отмены бронирования
CREATE TABLE cancellation_rules (
    id SERIAL PRIMARY KEY,
    rule_type VARCHAR(50) NOT NULL DEFAULT 'general' CHECK (rule_type IN ('general', 'refund', 'conditions')),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    display_order INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Создание индексов
CREATE INDEX idx_cancellation_rules_type_active ON cancellation_rules (rule_type, is_active);
CREATE INDEX idx_cancellation_rules_order ON cancellation_rules (display_order);

-- Добавление начальных правил отмены на основе текста с изображения
INSERT INTO cancellation_rules (rule_type, title, content, display_order) VALUES 
(
    'refund', 
    'Полный возврат', 
    'При отмене бронирования более чем за 24 часа до прибытия гость получает полный возврат уплаченной суммы.',
    1
),
(
    'refund',
    'Нет возврата', 
    'При отмене менее чем за 24 часа до прибытия.',
    2
),
(
    'conditions',
    'Подтверждение бронирования',
    'Бронирование подтверждается когда собственник примет запрос (в течение часа)',
    3
);

-- Комментарии для документации
COMMENT ON TABLE cancellation_rules IS 'Таблица правил отмены бронирований';
COMMENT ON COLUMN cancellation_rules.rule_type IS 'Тип правила: general - общие, refund - возврат средств, conditions - условия';
COMMENT ON COLUMN cancellation_rules.title IS 'Заголовок правила';
COMMENT ON COLUMN cancellation_rules.content IS 'Содержание правила';
COMMENT ON COLUMN cancellation_rules.is_active IS 'Активно ли правило';
COMMENT ON COLUMN cancellation_rules.display_order IS 'Порядок отображения'; 