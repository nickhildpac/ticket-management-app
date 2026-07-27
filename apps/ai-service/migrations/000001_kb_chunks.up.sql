-- pgvector-backed knowledge-base table for RAG retrieval.
--
-- Under ADR 0002 the Go ticket-service owns the ticket/user/comment schema;
-- this service's migrations only add AI-owned tables. The embedding dimension
-- must match EMBEDDING_DIM (config default 384); changing it needs a new
-- migration.
--
-- Every statement is IF NOT EXISTS so this applies cleanly to a database that
-- already has the table from the previous Alembic-managed revisions.
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS kb_chunks (
    id         SERIAL PRIMARY KEY,
    source     TEXT NOT NULL,
    content    TEXT NOT NULL,
    embedding  vector(384) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Approximate-nearest-neighbour index over cosine distance.
CREATE INDEX IF NOT EXISTS idx_kb_chunks_embedding
    ON kb_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
