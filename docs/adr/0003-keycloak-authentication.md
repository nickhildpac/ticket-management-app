# ADR 0003: Keycloak for authentication and roles

- **Status:** Accepted
- **Date:** 2026-08-18
- **Supersedes:** the authentication parts of ADR 0002 (service topology is unchanged)

## Context

Authentication was hand-rolled in the ticket-service: bcrypt password hashes in the `users` table,
self-signed HS256 access tokens, and a refresh token in an httpOnly cookie. The `JWT_SECRET` was
shared with the ai-service, which used it to mint tokens for the AI service account.

Two problems drove the change. First, the shared secret was an impersonation primitive: anything
holding it could mint a valid token for *any* user id, so a leak from the AI service compromised
every account. Second, the app owned credential storage, password reset, and session lifecycle —
security-critical surface with no reason to be bespoke.

## Decision

Delegate authentication to a **standalone Keycloak** deployment. Both Go services become pure OAuth2
resource servers: they hold no signing key, verify RS256 tokens against the realm's JWKS, and never
see a password.

### Deployment and validation

Keycloak runs as its own container (its own cluster in production). Token validation happens in
**service-level middleware only**. There is no gateway in this topology; adding one purely to
validate a second time would have bought nothing. The middleware is the single choke point, and
`internal/adapters/auth` is written so a gateway can be put in front later without code changes.

### Token strategy

Short-lived access tokens (5 minutes) plus refresh tokens, both from Keycloak's OIDC endpoints. The
SPA runs **Authorization Code + PKCE (S256)**. Implicit flow is disabled on every client: it returns
tokens in the URL fragment, where they leak into history and referrers, and PKCE removes any reason
to want it. Direct-grant (password) is disabled too, so no client can collect credentials itself.

The SPA refreshes at 75% of the access token's lifetime rather than waiting for a 401, with the 401
retry in `lib/api.ts` kept as a backstop for suspended tabs and clock skew. Logout is RP-initiated
against the realm's `end_session_endpoint`, which is what actually invalidates the refresh token and
SSO cookie server-side — clearing local state alone would let the next redirect sign the user
straight back in.

Tokens are held in memory, never `localStorage`. A reload therefore has no session and redirects
through Keycloak, which is invisible while the SSO cookie is valid. The alternative — persisting
tokens so reloads are instant — turns any XSS into a durable account takeover.

### Roles: realm roles, not Authorization Services

`admin` / `agent` / `user` are plain **realm roles**, mapped 1:1 onto `domain.UserRole`.
Authorization Services (policies, permissions, UMA tickets) was rejected: the rules that actually
matter here are row-dependent — *is this agent in this ticket's `assigned_to` array?* — and Keycloak
cannot evaluate those without replicating ticket data into it. So:

- **Keycloak decides role.** Route gating uses it via `RequireAnyRole` / `AdminRequired`.
- **Postgres decides ownership.** `internal/application/authorization` is unchanged.

`RoleFromRealmRoles` resolves multiple roles most-privileged-first and falls back to `user` when it
recognises none, so a misconfigured realm degrades to least privilege rather than granting admin.

### Identity mapping

Keycloak's `sub` is not usable as the local primary key: `tickets.created_by`,
`tickets.assigned_to[]` and `comments.created_by` are foreign keys into `users`, and repointing them
would rewrite history. Instead `users.keycloak_id` links the two (migration `000014`), and
`IdentityService` resolves a verified token to a local row by:

1. `keycloak_id` match — refreshing the row when the token's role or profile has drifted;
2. otherwise an **unlinked** row with the same email — claimed, so pre-Keycloak accounts keep their
   tickets;
3. otherwise JIT-provisioning a new row.

Step 2 refuses to link when the row already belongs to a different subject, which would otherwise
let a second account with a recycled email inherit someone else's tickets. Resolution is cached for
60s, but the cache is keyed on the token's role and profile as well as the subject, so a privilege
change in Keycloak applies on the next request rather than after the TTL.

### Service-to-service

The AI worker authenticates with **client_credentials** against its own confidential client, whose
service account holds the `admin` realm role. `JWT_SECRET` is gone from both services.

Role *writes* (the admin panel's role dropdown) go through the Keycloak Admin API using the
`ticket-service` client's service account, then mirror locally. Writing only locally would appear to
work and then be reverted on the user's next login; when admin credentials are absent the endpoint
returns 501 rather than pretending.

## Consequences

**Gained.** No shared signing secret and no impersonation primitive. No credential storage in the
app. Password reset, MFA, brute-force protection, and social/enterprise IdP federation become realm
configuration. Sessions are revocable centrally.

**Lost / costs.** Keycloak is now a hard runtime dependency for login — it is on the critical path
and needs its own HA story. Role management moves to Keycloak (or the Admin API path above).
Stored bcrypt hashes are deliberately destroyed by migration `000014`; the rollback restores the
schema but not the credentials, and the local login path no longer exists to use them.

**Removed.** `POST /api/v1/auth/login`, `/auth/logout`, `/auth/refresh`, `POST /api/v1/users`
(self-registration, now Keycloak's), `pkg/util/jwt_auth.go`, `pkg/util/password.go`, and the
refresh-cookie configuration. `GET /api/v1/auth/config` replaces them, publishing the issuer and
public client id so the SPA needs no per-environment rebuild.

**Not done.** No gateway, per the decision above. Skills remain local (they aren't identity). The
realm export ships dev users and dev client secrets, all of which are rejected by config validation
outside `local`/`test`.
