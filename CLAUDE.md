# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a ticket management system organized as a disciplined polyglot monorepo with three
deployable services, sharing a contracts boundary and one `make`-based entrypoint. **ADR 0002
(`docs/adr/0002-service-topology.md`) is the current architecture and supersedes ADR 0001.**

- **`apps/ticket-service/` (Go) — authoritative** for the ticket lifecycle: finite state machine,
  RBAC, authentication, tickets/comments/users persistence. This is the API of record.
- **`apps/ai-service/` (Go) — AI/RAG support triage.** A retrieval-augmented agent that decides
  whether a ticket can be **answered safely** or must be **escalated to a human**, then calls back
  into the ticket service. Ships three binaries: `cmd/api`, `cmd/worker`, `cmd/ingest`.
- **`apps/web/` (React + Vite, Bun)** — SPA, pointed at the Go ticket API.

- **Keycloak (standalone container)** — the identity provider. Authoritative for authentication and
  for the `admin`/`agent`/`user` realm roles. Both Go services are pure OAuth2 resource servers.
  See `docs/adr/0003-keycloak-authentication.md`.

Integration is event-driven: the Go service writes domain events to a transactional outbox, a relay
publishes them to Redis Streams, and the AI worker consumes them and applies triage results.

> Note: ADR 0001 previously made FastAPI authoritative and planned to remove the Go backend. That is
> reversed, and the ai-service has since been ported from Python/FastAPI to Go. There is no Python
> in this repo; do not add uv/pytest/ruff/Alembic tooling back.

## Development Commands

Prefer the root `Makefile` (single entrypoint): `make help` lists targets.

```bash
make setup        # install deps for all apps
make up           # core stack: postgres + keycloak + ticket-service + web  (profile: default)
make up-ai        # + redis + ai-service + worker                 (compose profile: ai)
make up-obs       # + prometheus + grafana                        (compose profile: obs)
make test         # go test (both services) + web build
make lint         # go vet (both services) + eslint
make contracts    # regenerate state-machine artifacts from contracts/
make migrate      # apply ai-service migrations (kb_chunks)
make ingest KB_PATH=knowledge   # ingest a knowledge base into the vector store
make worker       # run the AI triage worker (Redis Streams consumer)
make keycloak-reset             # re-import infra/keycloak/realm-export.json (drops console edits)
```

### Running backend services locally (without Docker)

Ports: **ticket-service 8080**, **ai-service 8081**, **keycloak 8090**, **postgres 5432**,
**redis 6379**. Bring up the datastores and Keycloak with Compose, then run each service from its app
directory. `.env` at the repo root holds shared values (Keycloak URLs/client secrets, DB creds,
`ANTHROPIC_API_KEY`).

```bash
# 1. Datastores + identity provider (redis only if you run the AI stack)
docker compose -f infra/compose/docker-compose.yml up -d postgres keycloak
docker compose -f infra/compose/docker-compose.yml --profile ai up -d redis
make migrate                                   # apply ai-service (kb_chunks) migrations

# 2. ticket-service (Go, authoritative API) — terminal 1
# KEYCLOAK_ISSUER_URL defaults to the local realm; leave KEYCLOAK_INTERNAL_ISSUER_URL
# unset outside Docker (localhost:8090 is reachable directly).
cd apps/ticket-service
DB_ADDR='postgres://postgres:postgres@localhost:5432/ticket_management?sslmode=disable' \
REDIS_ADDR='localhost:6379' PORT=8080 \
  go run cmd/api/main.go                        # REDIS_ADDR enables the outbox relay

# 3. ai-service API (triage + KB ingest endpoints) — terminal 2
cd apps/ai-service
go run ./cmd/api                                # :8081; also applies its own migrations

# 4. ai-service worker (consumes ticket-events, applies triage) — terminal 3
cd apps/ai-service
go run ./cmd/worker                             # needs ANTHROPIC_API_KEY + KEYCLOAK_CLIENT_ID/SECRET
```

The full stack via Compose is simpler: `make up` (core) or `make up-ai` (core + AI). Run services
directly (above) when you want live reload or a debugger attached.

### ticket-service (Go, authoritative) — `apps/ticket-service`
```bash
go run cmd/api/main.go         # run the API server (PORT default 8080)
go test ./...                  # all tests
go test ./internal/domain/...  # a single package
```
Config (`pkg/configs/loadconfig.go`): `PORT` (8080), `DB_ADDR` (a `postgres://` DSN), `APP_ENV`,
`KEYCLOAK_ISSUER_URL`, `KEYCLOAK_INTERNAL_ISSUER_URL`, `KEYCLOAK_AUDIENCE`, `KEYCLOAK_WEB_CLIENT_ID`,
and the optional `KEYCLOAK_ADMIN_CLIENT_ID`/`SECRET`. `REDIS_ADDR` enables the outbox relay. There is
no `JWT_SECRET`: the service verifies tokens against the realm's JWKS and signs nothing.

### ai-service (Go, RAG + triage) — `apps/ai-service`
```bash
go run ./cmd/api                 # /api/v1/triage, /api/v1/ingest, /health (PORT default 8081)
go run ./cmd/worker              # triage worker (Redis Streams consumer)
go run ./cmd/ingest knowledge    # chunk + embed a KB directory into pgvector
go test ./...                    # no database needed (Redis is faked with miniredis)
go vet ./...
```
Requires `ANTHROPIC_API_KEY` to run triage, and `KEYCLOAK_CLIENT_ID`/`KEYCLOAK_CLIENT_SECRET` for
callbacks into the ticket API. Settings live in `internal/config/config.go`; a local `.env` is loaded
automatically (without overriding already-set vars), and a SQLAlchemy-style `DATABASE_URL`
(`postgresql+psycopg://...`) is normalised automatically.

### web (React + Vite) — `apps/web`
```bash
bun run dev          # dev server (http://localhost:5173)
bun run build        # TypeScript check + Vite build
bun run lint
```
`VITE_API_URL` must point at the Go ticket-service (default `http://localhost:8080`). Keycloak's
issuer and client id are fetched at runtime from `GET /api/v1/auth/config`, so the bundle is
environment-independent; `VITE_KEYCLOAK_ISSUER` / `VITE_KEYCLOAK_CLIENT_ID` override that locally.

### Database Migrations
Both services use golang-migrate against the same database and each applies its own migrations on
startup. The ticket-service owns the ticket/user/comment schema (`apps/ticket-service/migrations/`,
tracked in `schema_migrations`); the ai-service owns only AI tables such as `kb_chunks`
(`apps/ai-service/migrations/`, tracked in **`ai_schema_migrations`** so the two histories don't
clobber each other).
```bash
make migrate                                     # or: cd apps/ai-service && go run ./cmd/api -migrate-only
```

## Architecture

### Technology Stack
- **ticket-service**: Go 1.24, Chi router, SQLC, PostgreSQL, hexagonal/clean layering
- **ai-service**: Go 1.24, Chi router, `database/sql` + pgvector, anthropic-sdk-go, go-redis
- **web**: React 19, TypeScript, Vite, TanStack Query/Router, Tailwind CSS v4
- **Orchestration**: Docker Compose profiles + root Makefile (no Turborepo/Nx)
- **Auth**: Keycloak (OIDC). SPA uses Authorization Code + PKCE; services validate RS256 via JWKS

### Project Structure
```
/
├── apps/
│   ├── ticket-service/   # Go — cmd/, internal/{domain,ports,application,adapters}, migrations/
│   │   └── internal/adapters/events/   # transactional outbox publisher + Redis relay
│   ├── ai-service/       # Go — cmd/{api,worker,ingest}, migrations/,
│   │   │                 #      internal/{config,rag,triage,worker,ticketapi,httpapi,dbmigrate,app}
│   └── web/              # Vite SPA — src/{app,components,features,lib}
├── contracts/
│   ├── ticket_state_machine.json   # state-machine source of truth
│   └── events/                     # async event schemas (ticket.created / ticket.updated)
├── infra/{compose,keycloak,monitoring}  # docker-compose + realm export + prometheus/grafana
├── tools/generate_ticket_state_contract.py   # codegen (runs gofmt on the Go artifact)
└── docs/adr/                       # 0001 (superseded), 0002 (current)
```

### Key Architecture Patterns

**ticket-service (Go, hexagonal):** routers/handlers in `internal/adapters/http`; workflows in
`internal/application/service`; ports in `internal/ports`; persistence in `internal/adapters/db`
(SQLC). `TicketService` emits `ticket.created`/`ticket.updated` via an `EventPublisher` port; the
`OutboxPublisher` writes to `event_outbox` and the `Relay` goroutine drains it to Redis.

**ai-service (RAG triage):** `internal/triage/agent.go` runs a tool loop over four tools
(`search_docs`, `rerank_results`, `draft_reply`, `escalate_ticket`) against KB context from
`internal/rag` (pgvector semantic + Postgres FTS, fused with RRF, then cross-encoder re-ranked). A
**deterministic safety gate** (`ApplySafetyGate`) decides auto-answer vs escalate. Fail-safe:
refusals, transport errors, an exhausted iteration budget and a loop with no terminal tool call all
escalate. `internal/worker` consumes the `ticket-events` Redis stream and applies results via
`internal/ticketapi`. Embeddings sit behind a swappable `Embedder` interface (default
`HashingEmbedder`; set `OPENAI_API_KEY` for real retrieval quality).

The agent drives `client.Messages.New` in a **manual tool loop** rather than the SDK's
`BetaToolRunner`: the runner defers tool execution to the start of the next turn, so a terminal
decision would cost one extra model round-trip before the loop could observe it and stop.

**web:** TanStack Router for routing, TanStack Query for server state; features under `src/features`,
shared API/domain helpers in `src/lib`. All API calls go to `VITE_API_URL` (the Go service).

## Domain Model

### Ticket State Machine

Enforced transitions; the shared contract in `contracts/ticket_state_machine.json` generates into all
both apps that enforce it (never hand-duplicate transition rules):

```
Open → Pending, In Progress, Cancelled
Pending → Open, In Progress, Resolved, Cancelled
In Progress → Resolved
Resolved → Open, Pending, Closed, Cancelled
Closed → (final state)
Cancelled → (final state)
```

Generated artifacts (the ai-service has no target — it never drives transitions):
- `apps/ticket-service/internal/domain/ticket_state_contract_gen.go` (Go)
- `apps/web/src/lib/ticket-state-contract.generated.ts` (TypeScript)

Run `make contracts` after editing the JSON. The generator runs `gofmt` on the Go output so
regeneration is byte-deterministic; the CI `contracts-drift` job fails on any drift.

### Core Entities

**Ticket**: `ID` (UUID), `CreatedBy` (UUID), `AssignedTo` (`uuid[]` in Go), `Skills`, `State`
(Open/Pending/InProgress/Resolved/Closed/Cancelled), `Priority` (Critical/High/Medium/Low),
`TicketNumber`. **Comment**: `ID`, `TicketID`, `CreatedBy`, `Description`. **User**: `ID`,
`KeycloakID` (nullable link to the Keycloak subject), `Email`, `FirstName`, `LastName`, `Role`.
Passwords are no longer stored — Keycloak owns credentials.

### Events

`contracts/events/ticket_events.json` defines the outbox → Redis envelope
(`event_id`/`event_type`/`aggregate_id`/`payload`) and the `TicketEventPayload`. Keep the Go
`domain.NewTicketEvent` payload and this schema in sync.

## API Structure (ticket-service, base `/api/v1`)

- `GET /tickets`, `GET /tickets?assigned_to=me`, `GET /tickets/{id}`, `POST /tickets`,
  `PATCH /tickets/{id}` (validates transitions), `DELETE /tickets/{id}`
- `POST /comments`, `GET /tickets/{id}/comments`
- Auth: `GET /auth/config` only (issuer + public client id for the SPA). Login, registration, token
  refresh and logout all happen against Keycloak directly — this service has no credential endpoints.

ai-service surface: `POST /api/v1/triage` (on-demand, returns a decision without applying it),
`POST /api/v1/ingest` (multipart KB upload; the ticket-service proxies `/api/v1/admin/documents` to
it behind AdminRequired), `GET /health` and `GET /api/v1/health`.

## Important Implementation Notes

### State Transition Validation
Always validate transitions through the generated contract helpers. The AI worker deliberately posts
comments only (valid from any state) and does not drive state transitions, to respect the FSM.

### AI safety boundary
The answer-safely-vs-escalate decision must remain enforced in code
(`internal/triage.Agent.ApplySafetyGate`), not left solely to the model: auto-answer only when the
model chose it, confidence ≥ threshold, no safety flags, and a non-empty draft. Refusals/errors
escalate.

### Authentication (Keycloak)
Both Go services are resource servers: they hold **no signing key** and verify RS256 tokens against
the realm's JWKS (`internal/adapters/auth` in ticket-service, `internal/keycloak` in ai-service).
Never reintroduce a shared `JWT_SECRET` or a local password path.

- **Roles** are Keycloak realm roles (`admin`/`agent`/`user`), reduced to one `domain.UserRole` by
  `RoleFromRealmRoles`, which falls back to the *least* privileged role on anything unrecognised.
  Route gating uses `RequireAnyRole`/`AdminRequired`; per-record ownership rules stay in
  `internal/application/authorization` because they depend on Postgres rows, not the token.
- **Identity**: Keycloak's `sub` is not the local user id. `users.keycloak_id` links them and
  `service.IdentityService` resolves a token to a local row (link by email → else JIT-provision),
  because tickets/comments have foreign keys into `users`. See ADR 0003.
- The SPA runs Authorization Code + PKCE against Keycloak directly (`apps/web/src/app/auth.ts`);
  tokens are in-memory only, never `localStorage`.

### AI → ticket-service callback auth
`internal/ticketapi` presents a Keycloak **client_credentials** token for the `ai-service` client,
whose service account holds the `admin` realm role — admin is required because only admins may
comment on tickets they aren't assigned to. The local row it owns is seeded by ticket-service
migration `000013` and linked to the pinned service-account subject in the realm export. The worker
calls `VerifyAccess` at startup and logs loudly if the callback path is misconfigured (non-fatal —
it still drains the stream).

### Type Safety
Both services and the web app are statically typed. Update `contracts/` then run `make contracts` —
do not edit generated files by hand.
