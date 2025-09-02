-- Create contracts table for storing rental contracts
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

-- Create index for faster lookups
CREATE INDEX idx_contracts_booking_id ON contracts(booking_id);
CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_created_at ON contracts(created_at);

-- Add comments
COMMENT ON TABLE contracts IS 'Rental contracts generated for confirmed bookings';
COMMENT ON COLUMN contracts.booking_id IS 'Reference to the booking this contract belongs to';
COMMENT ON COLUMN contracts.template_version IS 'Version of the contract template used';
COMMENT ON COLUMN contracts.generated_html IS 'Complete HTML content of the contract with substituted data';
COMMENT ON COLUMN contracts.status IS 'Contract status: draft, confirmed, signed'; 