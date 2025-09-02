-- Rollback добавления полей согласия в таблицу квартир
DROP INDEX IF EXISTS idx_apartments_agreement;
DROP INDEX IF EXISTS idx_apartments_contract;

ALTER TABLE apartments DROP CONSTRAINT IF EXISTS fk_apartments_contract_id;
ALTER TABLE apartments DROP COLUMN IF EXISTS contract_id;
ALTER TABLE apartments DROP COLUMN IF EXISTS agreement_accepted_at;
ALTER TABLE apartments DROP COLUMN IF EXISTS is_agreement_accepted; 