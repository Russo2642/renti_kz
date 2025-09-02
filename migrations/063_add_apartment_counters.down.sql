-- Remove indexes
DROP INDEX IF EXISTS idx_apartments_counters;
DROP INDEX IF EXISTS idx_apartments_booking_count;
DROP INDEX IF EXISTS idx_apartments_view_count;

-- Remove columns
ALTER TABLE apartments 
DROP COLUMN IF EXISTS booking_count,
DROP COLUMN IF EXISTS view_count;
