-- Transactional-outbox table for domain events consumed by the AI/RAG service.
-- See docs/adr/0002-service-topology.md (event-driven integration).
CREATE TABLE "event_outbox" (
  "id"           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "event_type"   varchar NOT NULL,
  "aggregate_id" UUID NOT NULL,
  "payload"      jsonb NOT NULL,
  "created_at"   timestamptz NOT NULL DEFAULT now(),
  "published_at" timestamptz
);

-- Partial index so the relay can cheaply poll only unpublished rows in order.
CREATE INDEX "idx_event_outbox_unpublished"
  ON "event_outbox" ("created_at")
  WHERE "published_at" IS NULL;
