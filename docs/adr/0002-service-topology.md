# ADR 0002: Service Topology — Go Tickets, Python AI/RAG, Vite Web

## Status

Accepted (2026-07-02). Supersedes [ADR 0001](0001-backend-direction.md).

## Context

ADR 0001 made the FastAPI service (`backend-python/`) the authoritative ticket-lifecycle backend
and planned to delete the Go service (`backend/`). That direction is being reversed, and a new
capability is being added: an AI agent that triages support tickets — deciding whether a ticket can
be **answered safely** or must be **escalated to a human** — backed by retrieval-augmented
generation (RAG) over a knowledge base of docs and resolved tickets.

## Decision

The system is three deployable services with one concern each, in a disciplined monorepo:

- **`apps/ticket-service/` (Go) — authoritative** for the ticket lifecycle: finite state machine,
  RBAC, authentication, tickets/comments/users persistence. This is the API of record; the web app
  and API clients target it.
- **`apps/ai-service/` (Python/FastAPI shell, repurposed)** — the AI/RAG support-triage service.
  Its ticket-CRUD code is stripped; it keeps the FastAPI app shell, config, and test harness and
  adds ingestion, retrieval, the triage agent, and an event worker. It reaches back into the ticket
  service's API to post replies or flag escalations.
- **`apps/web/` (Vite + React, unchanged framework)** — the existing SPA, relocated and repointed at
  the Go ticket API. It is deliberately **not** migrated to Next.js.

Cross-cutting decisions:

- **Shared contracts boundary (`contracts/`)** remains the single source of truth for the ticket
  state machine and, newly, async event schemas (`contracts/events/`). Language artifacts are
  generated, never hand-duplicated.
- **Integration is event-driven.** The ticket service writes domain events to a transactional
  outbox and a relay publishes them to a broker; the AI service consumes them, runs triage, and
  calls back into the ticket API. This decouples model/RAG latency from the request path.
- **Orchestration is Docker Compose profiles + a root Makefile + shell scripts** — intentionally
  lightweight for the repo's size (no Turborepo/Nx/Bazel).

## Consequences

- The Go service is now primary and actively developed; ADR 0001's plan to remove it is cancelled.
- Auth work that ADR 0001 targeted at FastAPI (hashed refresh-token rotation) is planned for the Go
  service but **not yet implemented** — the Go service still uses stateless refresh JWTs. Tracked as
  a follow-up.
- Go owns the tickets schema; the AI service owns only its own tables (e.g. the vector store).
- The FastAPI service no longer serves ticket CRUD; any lingering ticket-domain code there is debt
  to be removed as it is repurposed.
- Adds operational surface: a message broker and an AI worker. Kept minimal (at-least-once delivery
  + idempotency) with production hardening (DLQ, replay) left as follow-up.

## Migration Strategy

1. Reshape the repo into `apps/`, `infra/`, `tools/`; add the root Makefile and Compose profiles.
2. Make the Go service authoritative: port refresh-token rotation, add the transactional outbox.
3. Strip ticket CRUD from the FastAPI service; add ingest/rag/agent/worker.
4. Define event schemas in `contracts/events/` and generate Go + Python types; stand up the broker.
5. Repoint the web app at the Go API; update CI for three services + a contracts drift check.
