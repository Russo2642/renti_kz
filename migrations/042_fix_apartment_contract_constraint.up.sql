-- Fix apartment contract constraint to allow multiple rental contracts per apartment
-- but only one apartment contract per apartment

-- Drop the old constraint that was too restrictive
ALTER TABLE contracts DROP CONSTRAINT IF EXISTS unique_apartment_contract;

-- Add new constraint that only applies to apartment contracts
-- This allows multiple rental contracts per apartment but only one apartment contract
CREATE UNIQUE INDEX unique_apartment_contract_by_type 
ON contracts (apartment_id) 
WHERE type = 'apartment' AND is_active = true;

-- Add comment explaining the logic
COMMENT ON INDEX unique_apartment_contract_by_type IS 
'Ensures only one active apartment contract per apartment. Rental contracts can be multiple per apartment.'; 