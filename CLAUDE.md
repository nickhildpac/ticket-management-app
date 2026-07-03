# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a ticket management system organized as a disciplined polyglot monorepo with three
deployable services, sharing a contracts boundary and one `make`-based entrypoint. **ADR 0002
(`docs/adr/0002-service-topology.md`) is the current architecture and supersedes ADR 0001.**

- **`apps/ticket-service/` (Go) — authoritative** for the ticket lifecycle: finite state machine,
  RBAC, authentication, tickets/comments/users persistence. This is the API of record.
- **`apps/ai-service/` (Python/FastAPI) — AI/RAG support triage.** A retrieval-augmented agent that
  decides whether a ticket can be **answered safely** or must be **escalated to a human**, then calls
  back into the ticket service.
- **`apps/web/` (React + Vite, Bun)** — SPA, pointed at the Go ticket API.

Integration is event-driven: the Go service writes domain events to a transactional outbox, a relay
publishes them to Redis Streams, and the Python worker consumes them and applies triage results.

> Note: ADR 0001 previously made FastAPI authoritative and planned to remove the Go backend. That is
> reversed. Do not treat FastAPI as the ticket backend or the Go service as legacy.

## Development Commands

Prefer the root `Makefile` (single entrypoint): `make help` lists targets.

```bash
make setup        # install deps for all apps
make up           # core stack: postgres + ticket-service + web  (compose profile: default)
make up-ai        # + redis + ai-service                          (compose profile: ai)
make up-obs       # + prometheus + grafana                        (compose profile: obs)
make test         # go test + uv pytest + web build
make lint         # go vet + ruff + eslint
make contracts    # regenerate state-machine artifacts from contracts/
make migrate      # alembic upgrade head (ai-service tables)
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

# 3. ai-service API (FastAPI triage endpoint) — terminal 2
cd apps/ai-service
uv run uvicorn app.main:app --reload --port 8081

# 4. ai-service worker (consumes ticket-events, applies triage) — terminal 3
cd apps/ai-service
uv run python -m app.ai.worker                  # needs ANTHROPIC_API_KEY + a seeded AI_SERVICE_ACCOUNT_ID
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

### ai-service (Python/FastAPI, RAG + triage) — `apps/ai-service`
```bash
uv sync --extra dev
uv run uvicorn app.main:app --reload --port 8081   # /api/v1/triage + /health
uv run python -m app.ai.worker                     # triage worker
uv run pytest
uv run ruff check .
```
Requires `ANTHROPIC_API_KEY` to run triage. Key settings live in `app/core/config.py`.

### web (React + Vite) — `apps/web`
```bash
bun run dev          # dev server (http://localhost:5173)
bun run build        # TypeScript check + Vite build
bun run lint
```
`VITE_API_URL` must point at the Go ticket-service (default `http://localhost:8080`).

### Database Migrations
The Go ticket-service owns the ticket/user/comment schema via manual SQL migrations in
`apps/ticket-service/migrations/`. The ai-service's Alembic history (`apps/ai-service/alembic/`) now
only adds AI-owned tables (e.g. `kb_chunks`):
```bash
cd apps/ai-service && uv run alembic upgrade head   # or: make migrate
```

## Architecture

### Technology Stack
- **ticket-service**: Go 1.24, Chi router, SQLC, PostgreSQL, hexagonal/clean layering
- **ai-service**: FastAPI, SQLAlchemy, Alembic, pgvector, Anthropic SDK, Redis
- **web**: React 19, TypeScript, Vite, TanStack Query/Router, Tailwind CSS v4
- **Orchestration**: Docker Compose profiles + root Makefile (no Turborepo/Nx)
- **Auth**: JWT with cookie-based refresh tokens

### Project Structure
```
/
├── apps/
│   ├── ticket-service/   # Go — cmd/, internal/{domain,ports,application,adapters}, migrations/
│   │   └── internal/adapters/events/   # transactional outbox publisher + Redis relay
│   ├── ai-service/       # FastAPI — app/ai/ (embeddings, vectorstore, retrieval, agent, worker),
│   │   │                 #           app/api/triage.py, alembic/
│   │   └── app/api/v1/   # former ticket routers, retained for reference, NOT mounted
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

**ai-service (RAG triage):** `app/ai/agent.py` retrieves KB context (pgvector via
`vectorstore.py`/`retrieval.py`), calls Claude with `messages.parse` → a structured `TriageDecision`,
then a **deterministic safety gate** (`apply_safety_gate`) decides auto-answer vs escalate.
Fail-safe: refusals and errors escalate. `app/ai/worker.py` consumes the `ticket-events` Redis
stream and applies results via `ticket_client.py`. Embeddings are behind a swappable `Embedder`
protocol (default `HashingEmbedder`; replace with a real provider for quality retrieval).

**web:** TanStack Router for routing, TanStack Query for server state; features under `src/features`,
shared API/domain helpers in `src/lib`. All API calls go to `VITE_API_URL` (the Go service).

## Domain Model

### Ticket State Machine

Enforced transitions; the shared contract in `contracts/ticket_state_machine.json` generates into all
three languages (never hand-duplicate transition rules):

```
Open → Pending, In Progress, Cancelled
Pending → Open, In Progress, Resolved, Cancelled
In Progress → Resolved
Resolved → Open, Pending, Closed, Cancelled
Closed → (final state)
Cancelled → (final state)
```

Generated artifacts:
- `apps/ticket-service/internal/domain/ticket_state_contract_gen.go` (Go)
- `apps/ai-service/app/generated/ticket_state_contract.py` (Python)
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
`GET /health`.

## Important Implementation Notes

### State Transition Validation
Always validate transitions through the generated contract helpers. The AI worker deliberately posts
comments only (valid from any state) and does not drive state transitions, to respect the FSM.

### AI safety boundary
The answer-safely-vs-escalate decision must remain enforced in code (`apply_safety_gate`), not left
solely to the model: auto-answer only when the model chose it, confidence ≥ threshold, no safety
flags, and a non-empty draft. Refusals/errors escalate.

### AI → ticket-service callback auth
`app/ai/ticket_client.py` mints a short-lived JWT for the AI service account. It requires a seeded
agent/admin user id (`AI_SERVICE_ACCOUNT_ID`) and iss/aud matching the ticket-service middleware.

### Type Safety
Go and TypeScript are statically typed; FastAPI/Pydantic + SQLAlchemy define the ai-service contracts.
Update `contracts/` then run `make contracts` — do not edit generated files by hand.
