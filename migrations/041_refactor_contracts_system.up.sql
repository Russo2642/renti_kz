-- Refactor contracts system for new architecture
-- Drop old constraints and recreate table with new structure

-- Drop old table (we'll recreate with new structure)
DROP TABLE IF EXISTS contracts CASCADE;

-- Create new contracts table with enhanced structure
CREATE TABLE contracts (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL DEFAULT 'rental',
    apartment_id BIGINT NOT NULL,
    booking_id BIGINT NULL,
    template_version INTEGER NOT NULL DEFAULT 1,
    data_snapshot JSONB NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_contracts_apartment_id FOREIGN KEY (apartment_id) REFERENCES apartments(id) ON DELETE CASCADE,
    CONSTRAINT fk_contracts_booking_id FOREIGN KEY (booking_id) REFERENCES bookings(id) ON DELETE CASCADE,
    CONSTRAINT chk_contracts_type CHECK (type IN ('apartment', 'rental')),
    CONSTRAINT chk_contracts_status CHECK (status IN ('draft', 'confirmed', 'signed')),
    CONSTRAINT chk_contracts_booking_required CHECK (
        (type = 'rental' AND booking_id IS NOT NULL) OR 
        (type = 'apartment' AND booking_id IS NULL)
    ),
    CONSTRAINT unique_booking_contract UNIQUE (booking_id),
    CONSTRAINT unique_apartment_contract UNIQUE (apartment_id) 
        DEFERRABLE INITIALLY DEFERRED -- Allow temporary duplicates during migration
);

-- Create optimized indexes for fast lookups
CREATE INDEX idx_contracts_apartment_type ON contracts(apartment_id, type, is_active);
CREATE INDEX idx_contracts_booking_id ON contracts(booking_id) WHERE booking_id IS NOT NULL;
CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_created_at ON contracts(created_at);
CREATE INDEX idx_contracts_expires_at ON contracts(expires_at) WHERE expires_at IS NOT NULL;

-- GIN index for JSONB data_snapshot queries
CREATE INDEX idx_contracts_data_snapshot ON contracts USING GIN (data_snapshot);

-- Create template versions table for versioning
CREATE TABLE contract_templates (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,
    version INTEGER NOT NULL,
    template_content TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_template_type CHECK (type IN ('apartment', 'rental')),
    CONSTRAINT unique_type_version UNIQUE (type, version)
);

-- Insert initial template versions
INSERT INTO contract_templates (type, version, template_content, is_active) VALUES 
('rental', 1, 'rental_template_v1', true),
('apartment', 1, 'apartment_template_v1', true);

-- Add comments for documentation
COMMENT ON TABLE contracts IS 'Enhanced contracts system supporting apartment and rental contract types';
COMMENT ON COLUMN contracts.type IS 'Contract type: apartment (owner-company) or rental (owner-renter)';
COMMENT ON COLUMN contracts.apartment_id IS 'Reference to apartment (required for both types)';
COMMENT ON COLUMN contracts.booking_id IS 'Reference to booking (required only for rental type)';
COMMENT ON COLUMN contracts.data_snapshot IS 'JSONB snapshot of contract data at creation time';
COMMENT ON COLUMN contracts.is_active IS 'Whether the contract is currently active';
COMMENT ON COLUMN contracts.expires_at IS 'Contract expiration timestamp (NULL = no expiration)';

-- Create function for automatic updated_at
CREATE OR REPLACE FUNCTION update_contracts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for updated_at
CREATE TRIGGER trigger_contracts_updated_at
    BEFORE UPDATE ON contracts
    FOR EACH ROW
    EXECUTE FUNCTION update_contracts_updated_at(); 