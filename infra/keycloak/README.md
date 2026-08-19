# Keycloak realm

`realm-export.json` is imported on first boot (`start-dev --import-realm`). It is the source of
truth for the dev realm; edit it here rather than clicking through the admin console, or your
changes vanish the next time the Keycloak volume is recreated.

Admin console: <http://localhost:8090> (`admin` / `admin`, override with `KEYCLOAK_ADMIN*`).

## Why the realm looks the way it does

**Pinned user IDs.** Keycloak's `sub` is the identity the API trusts, but `tickets.created_by`,
`tickets.assigned_to[]` and `comments.created_by` are foreign keys into the local `users` table. The
seed users' Keycloak IDs are pinned to the same UUIDs as ticket-service migration `000009`, and the
`ai-service` service account to the UUID from migration `000013`. That makes `users.keycloak_id`
line up with `users.id` for everything that already exists, so seeded tickets keep their authors and
the AI triage account keeps its comment history.

Users created after this (self-registration, or anyone added in the console) get a fresh Keycloak
`sub` and are JIT-provisioned a *new* local row with its own `users.id`; only the seed rows are
special. See `internal/application/service/identity_service.go`.

**Realm roles, not Authorization Services.** `admin` / `agent` / `user` are plain realm roles, which
map 1:1 onto `domain.UserRole`. Ownership rules (can this agent see this ticket?) stay in
`internal/application/authorization` because they depend on `assigned_to` rows in Postgres — pushing
that into Keycloak would mean replicating ticket data into it.

`default-roles-ticket-management` composes `user`, so self-registered accounts land as end users
with no manual step. The Go role mapper also falls back to `user` if a token carries no recognised
realm role, so a misconfigured realm can never accidentally grant elevated access.

**Audience.** Keycloak does not add a resource-server audience by default, which would let a token
minted for any client in the realm hit the API. The `ticket-audience` client scope is a realm
default scope that stamps `aud: ticket-service`, and the Go verifier requires it.

**Clients.**

| Client | Type | Flow | Used by |
| --- | --- | --- | --- |
| `ticket-web` | public | Authorization Code + PKCE (S256) | the SPA |
| `ai-service` | confidential | `client_credentials` | triage worker callbacks |
| `ticket-service` | confidential | `client_credentials` | audience target, plus Admin API role writes |

Implicit and direct-grant (password) flows are disabled on every client on purpose.

The `ticket-service` service account holds the `realm-management` roles `view-users` and
`manage-users` so the app's admin panel can change a user's realm role. Drop
`KEYCLOAK_ADMIN_CLIENT_ID`/`SECRET` and that endpoint returns 501 instead — roles then change only
in this console.

## Issuer URL inside vs outside Docker

The browser reaches Keycloak at `http://localhost:8090`, but ticket-service reaches it at
`http://keycloak:8080` on the compose network. Tokens carry the *external* URL in `iss`, so the API
validates against `KEYCLOAK_ISSUER_URL` while fetching discovery/JWKS from
`KEYCLOAK_INTERNAL_ISSUER_URL`. When both are set and differ, the Go verifier uses go-oidc's
`InsecureIssuerURLContext` to allow exactly that split. Leave `KEYCLOAK_INTERNAL_ISSUER_URL` unset
when running the services outside Docker.

## Dev credentials

All seed users share the password `password123`.

| Email | Role |
| --- | --- |
| `alice@admin.com` | admin |
| `bob@agent.com`, `eve@agent.com` | agent |
| `charlie@user.com`, `diana@user.com`, `frank@user.com` | user |

Dev client secrets: `ai-service` → `ai-service-dev-secret`, `ticket-service` →
`ticket-service-dev-secret`. Both are rejected by config validation outside `local`/`test`.

## Editing the realm

Import happens on **first boot only**. After changing `realm-export.json`:

```bash
make keycloak-reset
```

That recreates the volume, so anything changed in the admin console is lost.

## Production notes

`start-dev` and the baked-in secrets/passwords here are for local development only. For a real
deployment: run `start` (not `start-dev`) with `KC_DB=postgres` against a dedicated database, put
Keycloak behind TLS with `KC_HOSTNAME` set to the public URL, replace both client secrets with
generated ones injected from your secret store, and drop the seeded users from the imported realm.
