-- Keycloak becomes the authentication authority (see docs/adr/0003-keycloak-authentication.md).
-- The local users table stays the identity of record for *ownership*: tickets.created_by,
-- tickets.assigned_to[] and comments.created_by are foreign keys into it. So rather than
-- repointing those at Keycloak subjects, we add the Keycloak subject as a linked identity.

ALTER TABLE users ADD COLUMN IF NOT EXISTS keycloak_id UUID;

-- One local row per Keycloak subject. NULLs are still allowed (and are not
-- compared by a unique index), which is what lets pre-existing rows sit
-- unlinked until someone signs in as them.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_keycloak_id ON users (keycloak_id);

-- Passwords are no longer stored or checked here. The column is kept rather than
-- dropped so existing rows (and the seed data) survive, but new rows created by
-- JIT provisioning have no password to supply.
ALTER TABLE users ALTER COLUMN hashed_password SET DEFAULT '!external';

-- Neutralise every stored bcrypt hash. Nothing in the service verifies passwords
-- any more, so leaving live hashes in the table is pure liability.
UPDATE users SET hashed_password = '!external' WHERE hashed_password <> '!disabled';

-- Link the accounts whose Keycloak IDs are pinned in infra/keycloak/realm-export.json
-- to the same UUIDs. That keeps seeded tickets/comments attributed to their authors,
-- and keeps the AI triage service account (migration 000013) owning its past comments.
-- Matching on email keeps this a no-op for any database that never ran the seeds.
UPDATE users SET keycloak_id = id
WHERE keycloak_id IS NULL
  AND id IN (
    'a1111111-1111-4111-8111-111111111111',
    'b2222222-2222-4222-8222-222222222222',
    'c3333333-3333-4333-8333-333333333333',
    'd4444444-4444-4444-8444-444444444444',
    'e5555555-5555-4555-8555-555555555555',
    'f6666666-6666-4666-8666-666666666666',
    '00000000-0000-4000-8000-0000000000a1'
  );
