-- Note: the bcrypt hashes overwritten by the up migration are not recoverable.
-- Rolling back restores the schema, not the credentials; re-seed if you need
-- local logins back (and that path no longer exists in the service anyway).
ALTER TABLE users ALTER COLUMN hashed_password DROP DEFAULT;

DROP INDEX IF EXISTS idx_users_keycloak_id;

ALTER TABLE users DROP COLUMN IF EXISTS keycloak_id;
