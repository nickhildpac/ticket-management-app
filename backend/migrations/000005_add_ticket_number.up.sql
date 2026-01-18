-- Create sequence for ticket numbers starting from 100001
CREATE SEQUENCE IF NOT EXISTS ticket_number_seq START 100001;

-- Add ticket_number column to tickets table
ALTER TABLE tickets ADD COLUMN ticket_number BIGINT UNIQUE NOT NULL DEFAULT nextval('ticket_number_seq');

-- Create index for fast lookup by ticket_number
CREATE INDEX idx_tickets_ticket_number ON tickets(ticket_number);

-- Set existing tickets' ticket_number if there are any
UPDATE tickets SET ticket_number = nextval('ticket_number_seq') WHERE ticket_number IS NULL;
