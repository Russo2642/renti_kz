-- Add view_count and booking_count to apartments table
ALTER TABLE apartments 
ADD COLUMN view_count INTEGER DEFAULT 0 NOT NULL,
ADD COLUMN booking_count INTEGER DEFAULT 0 NOT NULL;

-- Add indexes for performance when sorting by counters
CREATE INDEX idx_apartments_view_count ON apartments(view_count);
CREATE INDEX idx_apartments_booking_count ON apartments(booking_count);

-- Add composite index for sorting by both
CREATE INDEX idx_apartments_counters ON apartments(view_count, booking_count);
