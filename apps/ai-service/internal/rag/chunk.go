// Package rag holds the retrieval half of the triage pipeline: embeddings, the
// pgvector-backed knowledge-base store, hybrid fusion, cross-encoder re-ranking
// and knowledge-base ingestion.
package rag

// KBChunk is a retrieved knowledge-base passage with its similarity distance.
//
// ID identifies the row for hybrid fusion (RRF); it is zero for chunks that
// don't come from a table row (stubs, tests). Distance is lower-is-better —
// either a cosine distance or a monotone transform of the re-ranker score.
type KBChunk struct {
	ID       int     `json:"id,omitempty"`
	Content  string  `json:"content"`
	Source   string  `json:"source"`
	Distance float64 `json:"distance"`
}
