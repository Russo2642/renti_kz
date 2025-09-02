-- Rollback apartment contract constraint fix

-- Drop the new partial unique index
DROP INDEX IF EXISTS unique_apartment_contract_by_type;

-- Restore the old constraint (though it was problematic)
ALTER TABLE contracts ADD CONSTRAINT unique_apartment_contract UNIQUE (apartment_id) 
    DEFERRABLE INITIALLY DEFERRED; 