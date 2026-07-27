# Ticket Management App

A ticket management system built as a disciplined polyglot monorepo: three deployable services
sharing a contracts boundary and a single `make`-based entrypoint.

The current architecture is defined by **[ADR 0002](docs/adr/0002-service-topology.md)**, which
supersedes ADR 0001. See [`CLAUDE.md`](CLAUDE.md) for the full engineering guide and
[`DEPLOYMENT.md`](DEPLOYMENT.md) for deployment.

## Services

| Service | Stack | Role |
| --- | --- | --- |
| [`apps/ticket-service/`](apps/ticket-service) | Go 1.24, Chi, SQLC, PostgreSQL | **Authoritative API of record** — ticket lifecycle FSM, RBAC, auth, tickets/comments/users. |
| [`apps/ai-service/`](apps/ai-service) | Go 1.24, Chi, pgvector, anthropic-sdk-go | RAG support triage — decides whether a ticket can be **answered safely** or must be **escalated to a human**, then calls back into the ticket service. |
| [`apps/web/`](apps/web) | React 19, Vite, TanStack Query/Router, Tailwind v4 | SPA pointed at the Go ticket API. |

Integration is event-driven: the Go service writes domain events to a transactional outbox, a relay
publishes them to Redis Streams, and the AI worker consumes them and applies triage results.

## Quick start

Prerequisites: Docker + Compose, and (for local non-Docker runs) Go 1.24 and [Bun](https://bun.sh/).

```bash
make setup   # install deps for all apps
make up      # core stack: postgres + ticket-service + web
make up-ai   # + redis + ai-service + worker (AI triage)
make up-obs  # + prometheus + grafana
make down    # stop everything
```

- Web: http://localhost:5173
- ticket-service API: http://localhost:8080 (base path `/api/v1`)
- ai-service API: http://localhost:8081 (`/api/v1/triage`, `/api/v1/ingest`, `/health`)

## Common commands

Run `make help` for the full list. The root `Makefile` is the single entrypoint.

```bash
make test        # go test (both services) + web build
make lint         # go vet (both services) + eslint
make contracts   # regenerate state-machine artifacts from contracts/
make migrate     # apply ai-service (kb_chunks) migrations
make ingest KB_PATH=knowledge   # ingest a knowledge base into the vector store
make worker      # run the AI triage worker (Redis Streams consumer)
```

For live-reload / debugger workflows, run each service directly from its app directory — see the
"Running backend services locally" section of [`CLAUDE.md`](CLAUDE.md).

## Contracts

Cross-language sources of truth live in [`contracts/`](contracts):

- `ticket_state_machine.json` — the ticket state machine, generated into Go and TypeScript.
- `events/ticket_events.json` — the outbox → Redis event envelope.

Edit the JSON, then run `make contracts`. Never hand-edit generated artifacts; CI fails on drift.

### Ticket state machine

```
Open → Pending, In Progress, Cancelled
Pending → Open, In Progress, Resolved, Cancelled
In Progress → Resolved
Resolved → Open, Pending, Closed, Cancelled
Closed → (final)
Cancelled → (final)
```

## Repository layout

```
apps/
  ticket-service/   Go — cmd/, internal/{domain,ports,application,adapters}, migrations/
  ai-service/       Go — cmd/{api,worker,ingest}, internal/{rag,triage,worker,httpapi}, migrations/
  web/              Vite SPA — src/{app,components,features,lib}
contracts/          state-machine + async event schemas (source of truth)
infra/              docker-compose (profiles) + prometheus/grafana
tools/              contract codegen
docs/adr/           architecture decision records (0002 is current)
```

## Documentation

- [`CLAUDE.md`](CLAUDE.md) — architecture, patterns, and per-service development guide.
- [`DEPLOYMENT.md`](DEPLOYMENT.md) — deployment.
- [`docs/adr/`](docs/adr) — architecture decision records.
