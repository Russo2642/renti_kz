-- Добавление полей согласия с договором публикации в таблицу квартир
ALTER TABLE apartments ADD COLUMN is_agreement_accepted BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE apartments ADD COLUMN agreement_accepted_at TIMESTAMPTZ NULL;
ALTER TABLE apartments ADD COLUMN contract_id BIGINT NULL;

-- Добавление внешнего ключа на contracts
ALTER TABLE apartments ADD CONSTRAINT fk_apartments_contract_id 
FOREIGN KEY (contract_id) REFERENCES contracts(id) ON DELETE SET NULL;

-- Создание индекса для поиска по согласию
CREATE INDEX idx_apartments_agreement ON apartments(is_agreement_accepted);
CREATE INDEX idx_apartments_contract ON apartments(contract_id) WHERE contract_id IS NOT NULL;

-- Комментарии
COMMENT ON COLUMN apartments.is_agreement_accepted IS 'Принял ли владелец договор публикации на платформе';
COMMENT ON COLUMN apartments.agreement_accepted_at IS 'Время принятия договора';
COMMENT ON COLUMN apartments.contract_id IS 'Ссылка на договор между платформой и владельцем'; 