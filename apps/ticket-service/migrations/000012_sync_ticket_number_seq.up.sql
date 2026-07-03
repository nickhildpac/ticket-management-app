-- Seed data (000009) inserts explicit ticket_number values without advancing ticket_number_seq.
-- Sync the sequence so new tickets get MAX(ticket_number) + 1.
SELECT setval(
    'ticket_number_seq',
    GREATEST((SELECT COALESCE(MAX(ticket_number), 100000) FROM tickets), 100000),
    true
);
