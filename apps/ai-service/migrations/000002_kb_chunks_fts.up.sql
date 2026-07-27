-- GIN full-text index backing the keyword lane of hybrid retrieval
-- (semantic + FTS, fused with RRF). The expression must match the one in
-- VectorStore.SearchKeyword for the index to be used.
CREATE INDEX IF NOT EXISTS idx_kb_chunks_content_fts
    ON kb_chunks USING GIN (to_tsvector('english', content));
