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

Integration is event-driven: the Go service writes domain events to a transactional outbox, a relay
publishes them to Redis Streams, and the AI worker consumes them and applies triage results.

> Note: ADR 0001 previously made FastAPI authoritative and planned to remove the Go backend. That is
> reversed, and the ai-service has since been ported from Python/FastAPI to Go. There is no Python
> in this repo; do not add uv/pytest/ruff/Alembic tooling back.

## Development Commands

Prefer the root `Makefile` (single entrypoint): `make help` lists targets.

```bash
make setup        # install deps for all apps
make up           # core stack: postgres + ticket-service + web  (compose profile: default)
make up-ai        # + redis + ai-service + worker                 (compose profile: ai)
make up-obs       # + prometheus + grafana                        (compose profile: obs)
make test         # go test (both services) + web build
make lint         # go vet (both services) + eslint
make contracts    # regenerate state-machine artifacts from contracts/
make migrate      # apply ai-service migrations (kb_chunks)
make ingest KB_PATH=knowledge   # ingest a knowledge base into the vector store
make worker       # run the AI triage worker (Redis Streams consumer)
```

### Running backend services locally (without Docker)

Ports: **ticket-service 8080**, **ai-service 8081**, **postgres 5432**, **redis 6379**. Bring up the
datastores with Compose, then run each service from its app directory. `.env` at the repo root holds
shared values (`JWT_SECRET`, DB creds, `ANTHROPIC_API_KEY`, `AI_SERVICE_ACCOUNT_ID`).

```bash
# 1. Datastores (postgres always; redis only if you run the AI stack)
docker compose -f infra/compose/docker-compose.yml up -d postgres
docker compose -f infra/compose/docker-compose.yml --profile ai up -d redis
make migrate                                   # apply ai-service (kb_chunks) migrations

# 2. ticket-service (Go, authoritative API) — terminal 1
cd apps/ticket-service
DB_ADDR='postgres://postgres:postgres@localhost:5432/ticket_management?sslmode=disable' \
JWT_SECRET='local-dev-only-secret' REDIS_ADDR='localhost:6379' PORT=8080 \
  go run cmd/api/main.go                        # REDIS_ADDR enables the outbox relay

# 3. ai-service API (triage + KB ingest endpoints) — terminal 2
cd apps/ai-service
go run ./cmd/api                                # :8081; also applies its own migrations

# 4. ai-service worker (consumes ticket-events, applies triage) — terminal 3
cd apps/ai-service
go run ./cmd/worker                             # needs ANTHROPIC_API_KEY + a seeded AI_SERVICE_ACCOUNT_ID
```

The full stack via Compose is simpler: `make up` (core) or `make up-ai` (core + AI). Run services
directly (above) when you want live reload or a debugger attached.

### ticket-service (Go, authoritative) — `apps/ticket-service`
```bash
go run cmd/api/main.go         # run the API server (PORT default 8080)
go test ./...                  # all tests
go test ./internal/domain/...  # a single package
```
Config (`pkg/configs/loadconfig.go`): `PORT` (8080), `DB_ADDR` (a `postgres://` DSN), `JWT_SECRET`,
`APP_ENV`, cookie settings. `REDIS_ADDR` enables the outbox relay.

### ai-service (Go, RAG + triage) — `apps/ai-service`
```bash
go run ./cmd/api                 # /api/v1/triage, /api/v1/ingest, /health (PORT default 8081)
go run ./cmd/worker              # triage worker (Redis Streams consumer)
go run ./cmd/ingest knowledge    # chunk + embed a KB directory into pgvector
go test ./...                    # no database needed (Redis is faked with miniredis)
go vet ./...
```
Requires `ANTHROPIC_API_KEY` to run triage. Settings live in `internal/config/config.go`; every env
var keeps the name the Python service used, a local `.env` is loaded automatically (without
overriding already-set vars), and a SQLAlchemy-style `DATABASE_URL` (`postgresql+psycopg://...`) is
normalised automatically.

### web (React + Vite) — `apps/web`
```bash
bun run dev          # dev server (http://localhost:5173)
bun run build        # TypeScript check + Vite build
bun run lint
```
`VITE_API_URL` must point at the Go ticket-service (default `http://localhost:8080`).

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
- **Auth**: JWT with cookie-based refresh tokens

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
├── infra/{compose,monitoring}      # docker-compose (profiles) + prometheus/grafana
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
`TicketNumber`. **Comment**: `ID`, `TicketID`, `CreatedBy`, `Description`. **User**: `ID`, `Email`,
`HashedPassword`, `FirstName`, `LastName`, `Role`.

### Events

`contracts/events/ticket_events.json` defines the outbox → Redis envelope
(`event_id`/`event_type`/`aggregate_id`/`payload`) and the `TicketEventPayload`. Keep the Go
`domain.NewTicketEvent` payload and this schema in sync.

## API Structure (ticket-service, base `/api/v1`)

- `GET /tickets`, `GET /tickets?assigned_to=me`, `GET /tickets/{id}`, `POST /tickets`,
  `PATCH /tickets/{id}` (validates transitions), `DELETE /tickets/{id}`
- `POST /comments`, `GET /tickets/{id}/comments`
- Auth: `/auth/login`, `/auth/logout`, `/auth/refresh`

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

### AI → ticket-service callback auth
`internal/ticketapi` mints a short-lived JWT for the AI service account. A dedicated **admin**
service account is seeded by ticket-service migration `000013` (`AI_SERVICE_ACCOUNT_ID` defaults to
it); admin is required because only admins may comment on tickets they aren't assigned to. iss/aud
must match the ticket-service middleware. The worker calls `VerifyAccess` at startup and logs loudly
if the callback path is misconfigured (non-fatal — it still drains the stream).

### Type Safety
Both services and the web app are statically typed. Update `contracts/` then run `make contracts` —
do not edit generated files by hand.
