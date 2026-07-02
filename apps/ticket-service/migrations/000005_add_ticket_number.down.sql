-- Drop index
DROP INDEX IF EXISTS idx_tickets_ticket_number;

-- Drop column
ALTER TABLE tickets DROP COLUMN IF EXISTS ticket_number;

-- Drop sequence
DROP SEQUENCE IF EXISTS ticket_number_seq;
