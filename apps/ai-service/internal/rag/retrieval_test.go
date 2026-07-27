package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func chunk(id int, content string) KBChunk {
	return KBChunk{ID: id, Content: content, Source: "kb.md", Distance: 0.1}
}

func ids(chunks []KBChunk) []int {
	out := make([]int, len(chunks))
	for i, c := range chunks {
		out[i] = c.ID
	}
	return out
}

func TestRRFPrefersDocsRankedHighInBothLists(t *testing.T) {
	// Doc 1 is #1 in both -> highest RRF; doc 2 only in semantic; doc 3 only in keyword.
	semantic := []KBChunk{chunk(1, "a"), chunk(2, "b")}
	keyword := []KBChunk{chunk(1, "a"), chunk(3, "c")}

	fused := ReciprocalRankFusion([][]KBChunk{semantic, keyword}, 60)

	assert.Equal(t, []int{1, 2, 3}, ids(fused))
	// Doc 1's distance must be strictly better (lower) than the exclusive docs.
	assert.Less(t, fused[0].Distance, fused[1].Distance)
	assert.Less(t, fused[0].Distance, fused[2].Distance)
}

func TestRRFHandlesEmptyLanes(t *testing.T) {
	assert.Empty(t, ReciprocalRankFusion([][]KBChunk{{}, {}}, 60))

	onlySemantic := ReciprocalRankFusion([][]KBChunk{{chunk(9, "solo")}, {}}, 60)
	assert.Equal(t, []int{9}, ids(onlySemantic))
}

func TestRRFFallsBackToSourceContentKeyWithoutID(t *testing.T) {
	a := KBChunk{Content: "same", Source: "s", Distance: 0.2}
	b := KBChunk{Content: "same", Source: "s", Distance: 0.1}

	fused := ReciprocalRankFusion([][]KBChunk{{a}, {b}}, 60)

	require.Len(t, fused, 1)
	assert.Equal(t, "same", fused[0].Content)
}

func TestScoreToDistanceIsMonotoneInverted(t *testing.T) {
	assert.Less(t, ScoreToDistance(5.0), ScoreToDistance(0.0))
	assert.Less(t, ScoreToDistance(0.0), ScoreToDistance(-5.0))
}

func TestOpenRouterRerankerMapsScoresToInputOrder(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		// Ranked order differs from input order; the indices map back.
		_, _ = w.Write([]byte(`{"results":[
			{"index":1,"relevance_score":0.95},
			{"index":0,"relevance_score":0.4},
			{"index":2,"relevance_score":0.1}]}`))
	}))
	defer srv.Close()

	reranker := NewOpenRouterReranker("test-key", "cohere/rerank-v3.5", WithRerankBaseURL(srv.URL))
	scores, err := reranker.Score(context.Background(), "password reset", []string{"a", "b", "c"})

	require.NoError(t, err)
	assert.Equal(t, []float64{0.4, 0.95, 0.1}, scores)
	assert.Equal(t, "cohere/rerank-v3.5", got["model"])
	assert.Equal(t, float64(3), got["top_n"])
}

func TestOpenRouterRerankerRequiresAPIKey(t *testing.T) {
	reranker := NewOpenRouterReranker("", "cohere/rerank-v3.5")

	_, err := reranker.Score(context.Background(), "q", []string{"doc"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "OPENROUTER_API_KEY")
}

func TestOpenRouterRerankerSurfacesHTTPErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	reranker := NewOpenRouterReranker("k", "m", WithRerankBaseURL(srv.URL))
	_, err := reranker.Score(context.Background(), "q", []string{"doc"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

// stubStore serves fixed lanes so the fusion + re-rank pipeline runs end to end.
type stubStore struct {
	semantic []KBChunk
	keyword  []KBChunk
}

func (s *stubStore) SearchSemantic(_ context.Context, _ string, k int) ([]KBChunk, error) {
	return truncate(s.semantic, k), nil
}

func (s *stubStore) SearchKeyword(_ context.Context, _ string, k int) ([]KBChunk, error) {
	return truncate(s.keyword, k), nil
}

func truncate(chunks []KBChunk, k int) []KBChunk {
	if len(chunks) > k {
		return chunks[:k]
	}
	return chunks
}

// keywordReranker prefers "forgot password", then "reset link", then nothing.
type keywordReranker struct{}

func (keywordReranker) Score(_ context.Context, _ string, passages []string) ([]float64, error) {
	out := make([]float64, len(passages))
	for i, p := range passages {
		switch {
		case strings.Contains(p, "forgot password"):
			out[i] = 4.0
		case strings.Contains(p, "reset link"):
			out[i] = 2.0
		default:
			out[i] = -1.0
		}
	}
	return out, nil
}

func TestHybridRetrieverReranksToFinalTopK(t *testing.T) {
	store := &stubStore{
		semantic: []KBChunk{
			chunk(1, "password reset link"),
			chunk(2, "billing invoice pdf"),
			chunk(3, "forgot password steps"),
			chunk(4, "unrelated hardware tip"),
		},
		keyword: []KBChunk{
			chunk(3, "forgot password steps"),
			chunk(1, "password reset link"),
			chunk(5, "password manager export"),
		},
	}
	retriever := &HybridRetriever{
		Store:      store,
		Reranker:   keywordReranker{},
		CandidateK: 10,
		RerankPool: 5,
		RRFK:       60,
	}

	result, err := retriever.Retrieve(context.Background(), "I forgot my password", 3)

	require.NoError(t, err)
	require.Len(t, result, 3)
	assert.Equal(t, 3, result[0].ID)
	assert.Equal(t, 1, result[1].ID)
	assert.Less(t, result[0].Distance, result[1].Distance)
}

func TestFormatCandidatesLabelsByIDAndFallsBackToIndex(t *testing.T) {
	assert.Equal(t, "(no relevant knowledge-base passages were found)", FormatCandidates(nil))

	out := FormatCandidates([]KBChunk{
		{ID: 7, Source: "kb/a.md", Content: "body", Distance: 0.25},
		{Source: "kb/b.md", Content: "stub", Distance: 0.5},
	})

	assert.Contains(t, out, "[7] source=kb/a.md (distance=0.2500)")
	// A chunk with no id falls back to its 1-based position.
	assert.Contains(t, out, "[2] source=kb/b.md (distance=0.5000)")
}

func TestHashingEmbedderIsDeterministicAndNormalised(t *testing.T) {
	e := NewHashingEmbedder(16)

	a, err := e.Embed(context.Background(), "password reset")
	require.NoError(t, err)
	b, err := e.Embed(context.Background(), "password reset")
	require.NoError(t, err)

	assert.Equal(t, a, b)
	assert.Len(t, a, 16)

	var norm float64
	for _, v := range a {
		norm += float64(v) * float64(v)
	}
	assert.InDelta(t, 1.0, norm, 1e-6)
}

func TestBuildEmbedderPicksOpenAIWhenKeySet(t *testing.T) {
	offline := BuildEmbedder(EmbedderSettings{EmbeddingDim: 384})
	assert.IsType(t, &HashingEmbedder{}, offline)

	semantic := BuildEmbedder(EmbedderSettings{
		OpenAIAPIKey: "sk-test", EmbeddingModel: "text-embedding-3-small", EmbeddingDim: 384,
	})
	assert.IsType(t, &OpenAIEmbedder{}, semantic)
	assert.Equal(t, 384, semantic.Dim())
}
