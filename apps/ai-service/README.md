# ai-service (Go)

AI/RAG support-triage service. A retrieval-augmented agent decides whether a
ticket can be **answered safely** from the knowledge base or must be
**escalated to a human**, then calls back into the authoritative Go
ticket-service. See `docs/adr/0002-service-topology.md`.

This service does **not** own the ticket lifecycle — it only comments on
tickets, which is valid from any state, so it can't violate the ticket state
machine.

## Binaries

| Command | What it does |
| --- | --- |
| `cmd/api` | HTTP API: `POST /api/v1/triage`, `POST /api/v1/ingest`, `GET /health`. Applies migrations on startup. |
| `cmd/worker` | Consumes the `ticket-events` Redis stream, triages, and comments back. **This is the primary path.** |
| `cmd/ingest` | CLI that chunks and embeds a knowledge-base directory into pgvector. |

## Run locally

Postgres (with pgvector) and Redis come from the root compose file:

```bash
docker compose -f ../../infra/compose/docker-compose.yml up -d postgres
docker compose -f ../../infra/compose/docker-compose.yml --profile ai up -d redis
```

Then, from this directory:

```bash
cp .env.example .env
go run ./cmd/api      # :8081 — also applies the kb_chunks migrations
go run ./cmd/worker   # needs ANTHROPIC_API_KEY + a seeded AI_SERVICE_ACCOUNT_ID
go run ./cmd/ingest knowledge
```

`go test ./...` runs the suite; it needs no database (Postgres access is
covered by the ticket-service integration tests, and Redis is faked in-process
with miniredis).

## Architecture

```
internal/
  config/      env parsing + JWT-secret validation
  rag/         embeddings, pgvector store, RRF fusion, re-ranking, chunking
  triage/      the agent: system prompt, four tools, deterministic safety gate
  worker/      Redis Streams consumer (idempotency, reclaim, dead-letter)
  ticketapi/   callback client into the Go ticket API (mints a service JWT)
  httpapi/     chi router, shared-secret auth, triage + ingest handlers
  dbmigrate/   golang-migrate runner for this service's own tables
  app/         shared wiring for the three binaries
```

### The safety boundary

`triage.Agent.ApplySafetyGate` is deterministic and has the final say. The
model's `draft_reply` tool only *proposes* an auto-answer; the gate approves it
only when the model chose it, confidence ≥ threshold, no safety flags were
raised, and the draft is non-empty. Refusals, transport errors, an exhausted
iteration budget, and a loop that ends with no terminal tool call all escalate.

### Why a manual tool loop

The agent drives `client.Messages.New` itself rather than using the SDK's
`BetaToolRunner`. The runner defers tool execution to the start of the next
turn, so a terminal decision would cost one extra model round-trip before the
loop could observe it and stop.

## Migrations

This service owns only AI tables (`kb_chunks` and its indexes), applied with
golang-migrate from `migrations/`. It tracks them in **`ai_schema_migrations`**,
not golang-migrate's default `schema_migrations`, because the ticket-service
applies its own migrations into that table in the same database.

The SQL is written with `IF NOT EXISTS` throughout, so it also applies cleanly
to a database whose `kb_chunks` table was created by the previous
Alembic-managed revisions.

## Env vars

See `.env.example`. A `.env` in the working directory is loaded automatically
(it never overrides an already-set variable, so compose's `environment:` still
wins). A `DATABASE_URL` carried over from the Python service
(`postgresql+psycopg://...`) is normalised automatically, and `sslmode=disable`
is defaulted when absent.
