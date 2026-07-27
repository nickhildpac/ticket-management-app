package rag

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	"github.com/pgvector/pgvector-go"
)

// VectorStore is the pgvector-backed knowledge-base store living in the shared
// Postgres.
//
// The kb_chunks table is created by this service's migrations. This type only
// reads/writes rows and runs semantic (cosine) and keyword (FTS) search.
type VectorStore struct {
	db       *sql.DB
	embedder Embedder
}

// NewVectorStore wires a store to a database handle and an embedder. The
// embedder's Dim must match the kb_chunks.embedding column width.
func NewVectorStore(db *sql.DB, embedder Embedder) *VectorStore {
	return &VectorStore{db: db, embedder: embedder}
}

// Embedder exposes the configured embedder (ingestion reuses it).
func (s *VectorStore) Embedder() Embedder { return s.embedder }

// Add embeds content and stores it as a knowledge-base chunk.
func (s *VectorStore) Add(ctx context.Context, source, content string) error {
	embedding, err := s.embedder.Embed(ctx, content)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO kb_chunks (source, content, embedding) VALUES ($1, $2, $3)`,
		source, content, pgvector.NewVector(embedding))
	return err
}

// Search is a backward-compatible alias for semantic search.
func (s *VectorStore) Search(ctx context.Context, query string, k int) ([]KBChunk, error) {
	return s.SearchSemantic(ctx, query, k)
}

// SearchSemantic ranks chunks by cosine distance to the query embedding.
func (s *VectorStore) SearchSemantic(ctx context.Context, query string, k int) ([]KBChunk, error) {
	embedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}
	const stmt = `
		SELECT id, content, source, embedding <=> $1 AS distance
		FROM kb_chunks
		ORDER BY embedding <=> $1
		LIMIT $2`
	rows, err := s.db.QueryContext(ctx, stmt, pgvector.NewVector(embedding), k)
	if err != nil {
		return nil, err
	}
	return scanChunks(rows)
}

// SearchKeyword runs Postgres full-text search ranked by ts_rank_cd.
//
// Terms are OR-ed: plainto_tsquery ANDs every lexeme, so a multi-sentence
// ticket (title + description) matched nothing unless a single chunk contained
// *all* its words. OR-ing lets ts_rank_cd reward chunks that hit the most terms
// instead.
//
// Returns no rows when the query has no useful terms. Distance is set to
// 1 - rank (clipped at 0) so lower-is-better still holds.
func (s *VectorStore) SearchKeyword(ctx context.Context, query string, k int) ([]KBChunk, error) {
	terms := tsqueryOrTerms(query)
	if terms == "" {
		return nil, nil
	}
	// A single statement so the GIN expression index can be used. Rank is
	// higher-is-better; we invert it for KBChunk.Distance consistency.
	const stmt = `
		SELECT id, content, source,
		       1 - ts_rank_cd(to_tsvector('english', content), query) AS distance
		FROM kb_chunks,
		     to_tsquery('english', $1) AS query
		WHERE to_tsvector('english', content) @@ query
		ORDER BY ts_rank_cd(to_tsvector('english', content), query) DESC
		LIMIT $2`
	rows, err := s.db.QueryContext(ctx, stmt, terms, k)
	if err != nil {
		return nil, err
	}
	chunks, err := scanChunks(rows)
	if err != nil {
		return nil, err
	}
	for i := range chunks {
		if chunks[i].Distance < 0 {
			chunks[i].Distance = 0
		}
	}
	return chunks, nil
}

func scanChunks(rows *sql.Rows) ([]KBChunk, error) {
	defer rows.Close()
	var out []KBChunk
	for rows.Next() {
		var c KBChunk
		if err := rows.Scan(&c.ID, &c.Content, &c.Source, &c.Distance); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

var termRE = regexp.MustCompile(`[a-z0-9]+`)

// tsqueryOrTerms builds an OR-ed to_tsquery input from free text.
//
// Tokens are restricted to alphanumerics so user text can't inject tsquery
// syntax; duplicates are dropped (order-preserving). Stopwords are left for
// to_tsquery('english', ...) to discard.
func tsqueryOrTerms(query string) string {
	seen := make(map[string]struct{})
	var terms []string
	for _, token := range termRE.FindAllString(strings.ToLower(query), -1) {
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		terms = append(terms, token)
	}
	return strings.Join(terms, " | ")
}
