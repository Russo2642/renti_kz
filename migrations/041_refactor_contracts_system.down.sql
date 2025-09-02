-- Rollback refactor contracts system

-- Drop new tables and functions
DROP TRIGGER IF EXISTS trigger_contracts_updated_at ON contracts;
DROP FUNCTION IF EXISTS update_contracts_updated_at();
DROP TABLE IF EXISTS contract_templates CASCADE;
DROP TABLE IF EXISTS contracts CASCADE;

-- Recreate old contracts table structure
CREATE TABLE contracts (
    id BIGSERIAL PRIMARY KEY,
    booking_id BIGINT NOT NULL,
    template_version INTEGER NOT NULL DEFAULT 1,
    generated_html TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_contracts_booking_id FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE,
    CONSTRAINT chk_contracts_status CHECK (status IN ('draft', 'confirmed', 'signed')),
    CONSTRAINT unique_booking_contract UNIQUE (booking_id)
);

-- Recreate old indexes
CREATE INDEX idx_contracts_booking_id ON contracts(booking_id);
CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_created_at ON contracts(created_at); 