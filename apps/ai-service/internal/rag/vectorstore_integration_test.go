package rag

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// itestPrefix scopes this file's fixtures. The target database usually already
// holds a real ingested knowledge base, so assertions compare ranks *among the
// fixtures* rather than assuming an empty table.
const itestPrefix = "itest/"

// These exercise the SQL the unit tests can't reach: the pgvector cosine
// operator, the ts_rank_cd keyword lane, and the driver round-trip for the
// vector column. Gated on TEST_DATABASE_URL, matching the ticket-service's
// integration test.
//
//	TEST_DATABASE_URL='postgres://postgres:postgres@localhost:5432/ticket_management?sslmode=disable' go test ./internal/rag/
func newIntegrationStore(t *testing.T) (*VectorStore, context.Context) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	require.NoError(t, db.PingContext(ctx))

	// The offline embedder keeps the test hermetic — no API key required.
	store := NewVectorStore(db, NewHashingEmbedder(384))

	cleanup := func() {
		_, _ = db.ExecContext(context.Background(),
			`DELETE FROM kb_chunks WHERE source LIKE $1`, itestPrefix+"%")
	}
	cleanup() // drop leftovers from a previous failed run
	t.Cleanup(cleanup)

	return store, ctx
}

// fixtureSources returns the sources of this test's own rows, in result order.
func fixtureSources(chunks []KBChunk) []string {
	var out []string
	for _, c := range chunks {
		if strings.HasPrefix(c.Source, itestPrefix) {
			out = append(out, c.Source)
		}
	}
	return out
}

func TestVectorStoreSemanticSearchRoundTrip(t *testing.T) {
	store, ctx := newIntegrationStore(t)

	const (
		passwordText = "Use the reset link on the login page."
		billingText  = "Invoices are issued monthly in arrears."
	)
	require.NoError(t, store.Add(ctx, itestPrefix+"password.md", passwordText))
	require.NoError(t, store.Add(ctx, itestPrefix+"billing.md", billingText))

	// Querying with a passage's own text must surface that passage — the
	// embedding, the vector column round-trip and the cosine operator all have
	// to line up for this to hold. The target database usually also holds a
	// real knowledge base, so this asserts retrieval, not global rank.
	for _, tc := range []struct{ query, wantSource string }{
		{passwordText, itestPrefix + "password.md"},
		{billingText, itestPrefix + "billing.md"},
	} {
		got, err := store.SearchSemantic(ctx, tc.query, 5)
		require.NoError(t, err)
		require.NotEmpty(t, got)

		assert.Contains(t, fixtureSources(got), tc.wantSource)
		assert.True(t, sortedAscending(got), "semantic results should be ordered by distance")
		for _, c := range got {
			assert.NotZero(t, c.ID)
			// Cosine distance is lower-is-better and bounded to [0, 2].
			assert.GreaterOrEqual(t, c.Distance, 0.0)
			assert.LessOrEqual(t, c.Distance, 2.0)
		}
	}
}

func sortedAscending(chunks []KBChunk) bool {
	for i := 1; i < len(chunks); i++ {
		if chunks[i].Distance < chunks[i-1].Distance {
			return false
		}
	}
	return true
}

func TestVectorStoreKeywordSearchRanksByTermOverlap(t *testing.T) {
	store, ctx := newIntegrationStore(t)

	require.NoError(t, store.Add(ctx, itestPrefix+"password.md",
		"Password reset requires the login page link."))
	require.NoError(t, store.Add(ctx, itestPrefix+"hardware.md",
		"Replace the keyboard when keys stick."))

	got, err := store.SearchKeyword(ctx, "password reset login", 2000)

	require.NoError(t, err)
	sources := fixtureSources(got)
	require.NotEmpty(t, sources)
	assert.Equal(t, itestPrefix+"password.md", sources[0],
		"the passage sharing terms with the query must rank ahead of the unrelated one")

	for _, c := range got {
		// Distance is 1 - rank, clipped at 0, so it stays lower-is-better.
		assert.GreaterOrEqual(t, c.Distance, 0.0)
	}
}

func TestVectorStoreKeywordSearchIgnoresQueriesWithoutTerms(t *testing.T) {
	store, ctx := newIntegrationStore(t)

	got, err := store.SearchKeyword(ctx, "!!! ???", 5)

	require.NoError(t, err)
	assert.Empty(t, got)
}

// Free-text must never be able to inject tsquery syntax — the tokenizer strips
// the operators rather than letting Postgres parse them.
func TestVectorStoreKeywordSearchSurvivesTsqueryMetacharacters(t *testing.T) {
	store, ctx := newIntegrationStore(t)

	require.NoError(t, store.Add(ctx, itestPrefix+"password.md", "Password reset instructions."))

	got, err := store.SearchKeyword(ctx, "password & | ! ( ) <-> reset", 2000)

	require.NoError(t, err)
	assert.NotEmpty(t, fixtureSources(got), "the query must still match rather than erroring")
}

func TestIngestDocumentStoresChunks(t *testing.T) {
	store, ctx := newIntegrationStore(t)

	added, reason, err := IngestDocument(ctx, store, itestPrefix+"kb.md",
		[]byte("# One\n\nfirst body\n\n# Two\n\nsecond body"))

	require.NoError(t, err)
	assert.Equal(t, SkipNone, reason)
	assert.Equal(t, 2, added, "heading-aware chunking should produce one chunk per section")

	got, err := store.SearchKeyword(ctx, "second body", 2000)
	require.NoError(t, err)
	assert.NotEmpty(t, fixtureSources(got))
}

func TestIngestDocumentSkipsBinaryAndEmptyFiles(t *testing.T) {
	store, ctx := newIntegrationStore(t)

	added, reason, err := IngestDocument(ctx, store, itestPrefix+"logo.png", []byte("PNG\x00\x01"))
	require.NoError(t, err)
	assert.Zero(t, added)
	assert.Equal(t, SkipBinary, reason)

	added, reason, err = IngestDocument(ctx, store, itestPrefix+"blank.md", []byte("   \n\n  "))
	require.NoError(t, err)
	assert.Zero(t, added)
	assert.Equal(t, SkipEmpty, reason)
}
