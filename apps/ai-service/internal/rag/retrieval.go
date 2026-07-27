package rag

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// fusionKey identifies a chunk across ranked lists. Rows are keyed by their id;
// chunks without one (stubs, tests) fall back to source+content.
type fusionKey struct {
	id      int
	source  string
	content string
}

func keyFor(c KBChunk) fusionKey {
	if c.ID != 0 {
		return fusionKey{id: c.ID}
	}
	return fusionKey{source: c.Source, content: c.Content}
}

// ReciprocalRankFusion fuses multiple ranked lists with Reciprocal Rank Fusion.
//
//	score(doc) = Σ 1 / (rrfK + rank)
//
// over the lists where the doc appears (1-based ranks). Returns chunks sorted
// by RRF score descending — ties keep first-seen order — with Distance set to
// 1 / (1 + score) so lower-is-better still holds.
func ReciprocalRankFusion(rankedLists [][]KBChunk, rrfK int) []KBChunk {
	scores := make(map[fusionKey]float64)
	best := make(map[fusionKey]KBChunk)
	order := make([]fusionKey, 0)

	for _, ranked := range rankedLists {
		for i, chunk := range ranked {
			rank := i + 1
			key := keyFor(chunk)
			if _, seen := scores[key]; !seen {
				order = append(order, key)
			}
			scores[key] += 1.0 / float64(rrfK+rank)
			// Prefer a copy that carries an id when available.
			if prev, ok := best[key]; !ok || (prev.ID == 0 && chunk.ID != 0) {
				best[key] = chunk
			}
		}
	}

	sort.SliceStable(order, func(i, j int) bool {
		return scores[order[i]] > scores[order[j]]
	})

	out := make([]KBChunk, 0, len(order))
	for _, key := range order {
		chunk := best[key]
		out = append(out, KBChunk{
			ID:       chunk.ID,
			Content:  chunk.Content,
			Source:   chunk.Source,
			Distance: 1.0 / (1.0 + scores[key]),
		})
	}
	return out
}

// Retriever fetches knowledge-base passages relevant to a free-text query.
type Retriever interface {
	Retrieve(ctx context.Context, query string, k int) ([]KBChunk, error)
}

// VectorRetriever retrieves knowledge-base passages via pure semantic (vector)
// search.
//
// A convenience wrapper for tests and simple callers. The triage agent instead
// drives the underlying primitives (SearchSemantic/SearchKeyword +
// ReciprocalRankFusion + the re-ranker) from its search_docs and
// rerank_results tools.
type VectorRetriever struct{ Store *VectorStore }

func (v *VectorRetriever) Retrieve(ctx context.Context, query string, k int) ([]KBChunk, error) {
	return v.Store.SearchSemantic(ctx, query, k)
}

// HybridRetriever runs semantic + keyword retrieval, fuses with RRF, then
// re-ranks with a cross-encoder.
//
// Each lane fetches CandidateK hits; RRF merges them; the top RerankPool fused
// docs are scored by the re-ranker; the best k are returned. Kept as a one-shot
// retriever for tests and simple callers; the triage agent runs the same
// pipeline in stages across its tool calls.
type HybridRetriever struct {
	Store      SearchStore
	Reranker   Reranker
	CandidateK int
	RerankPool int
	RRFK       int
}

// SearchStore is the retrieval surface HybridRetriever needs (satisfied by
// *VectorStore; stubbed in tests).
type SearchStore interface {
	SearchSemantic(ctx context.Context, query string, k int) ([]KBChunk, error)
	SearchKeyword(ctx context.Context, query string, k int) ([]KBChunk, error)
}

func (h *HybridRetriever) Retrieve(ctx context.Context, query string, k int) ([]KBChunk, error) {
	semantic, err := h.Store.SearchSemantic(ctx, query, h.CandidateK)
	if err != nil {
		return nil, err
	}
	keyword, err := h.Store.SearchKeyword(ctx, query, h.CandidateK)
	if err != nil {
		return nil, err
	}
	fused := ReciprocalRankFusion([][]KBChunk{semantic, keyword}, h.RRFK)
	if len(fused) == 0 {
		return nil, nil
	}

	pool := fused
	if len(pool) > h.RerankPool {
		pool = pool[:h.RerankPool]
	}
	scored, err := h.rerank(ctx, query, pool)
	if err != nil {
		return nil, err
	}
	if len(scored) > k {
		scored = scored[:k]
	}
	return scored, nil
}

// rerank scores pool against query and returns it best-first, with Distance
// rewritten from the re-ranker score.
func (h *HybridRetriever) rerank(ctx context.Context, query string, pool []KBChunk) ([]KBChunk, error) {
	passages := make([]string, len(pool))
	for i, c := range pool {
		passages[i] = c.Content
	}
	scores, err := h.Reranker.Score(ctx, query, passages)
	if err != nil {
		return nil, err
	}
	if len(scores) != len(pool) {
		return nil, fmt.Errorf("reranker returned %d scores for %d passages", len(scores), len(pool))
	}

	idx := make([]int, len(pool))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return scores[idx[a]] > scores[idx[b]] })

	out := make([]KBChunk, len(pool))
	for rank, i := range idx {
		out[rank] = KBChunk{
			ID:       pool[i].ID,
			Content:  pool[i].Content,
			Source:   pool[i].Source,
			Distance: ScoreToDistance(scores[i]),
		}
	}
	return out, nil
}

const noPassages = "(no relevant knowledge-base passages were found)"

// FormatContext renders retrieved chunks into a grounded context block for the
// prompt.
func FormatContext(chunks []KBChunk) string {
	if len(chunks) == 0 {
		return noPassages
	}
	parts := make([]string, len(chunks))
	for i, chunk := range chunks {
		parts[i] = fmt.Sprintf("[%d] source=%s\n%s", i+1, chunk.Source, chunk.Content)
	}
	return strings.Join(parts, "\n\n")
}

// FormatCandidates renders chunks for a tool result the model can cite.
//
// Each passage is labelled with its KB id (so the model can cite "[id]" and
// rerank_results can reference the same candidates); chunks without an id
// (e.g. stubs) fall back to a 1-based index.
func FormatCandidates(chunks []KBChunk) string {
	if len(chunks) == 0 {
		return noPassages
	}
	parts := make([]string, len(chunks))
	for i, chunk := range chunks {
		label := chunk.ID
		if label == 0 {
			label = i + 1
		}
		parts[i] = fmt.Sprintf("[%d] source=%s (distance=%.4f)\n%s",
			label, chunk.Source, chunk.Distance, chunk.Content)
	}
	return strings.Join(parts, "\n\n")
}
