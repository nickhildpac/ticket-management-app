package rag

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var tokenRE = regexp.MustCompile(`[a-z0-9]+`)

// Embedder turns text into a fixed-length vector. Swap the implementation for a
// different embedding provider (Voyage AI, sentence-transformers, ...) by
// satisfying this interface.
type Embedder interface {
	// Dim is the vector length. It must match the kb_chunks.embedding column.
	Dim() int
	Embed(ctx context.Context, text string) ([]float32, error)
}

// HashingEmbedder is a deterministic, dependency-free bag-of-words hashing
// embedder.
//
// Good enough to exercise the RAG plumbing offline and in tests. It is *not* a
// semantic model — replace it with a real embedding provider for quality
// retrieval. Kept behind the Embedder interface so the swap is one line.
type HashingEmbedder struct{ dim int }

// NewHashingEmbedder builds an offline embedder producing dim-length vectors.
func NewHashingEmbedder(dim int) *HashingEmbedder { return &HashingEmbedder{dim: dim} }

func (h *HashingEmbedder) Dim() int { return h.dim }

func (h *HashingEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	vec := make([]float32, h.dim)
	for _, token := range tokenRE.FindAllString(strings.ToLower(text), -1) {
		digest := md5.Sum([]byte(token))
		bucket := binary.BigEndian.Uint32(digest[:4]) % uint32(h.dim)
		sign := float32(-1)
		if digest[4]&1 == 1 {
			sign = 1
		}
		vec[bucket] += sign
	}
	var sumSquares float64
	for _, v := range vec {
		sumSquares += float64(v) * float64(v)
	}
	if norm := math.Sqrt(sumSquares); norm > 0 {
		for i := range vec {
			vec[i] = float32(float64(vec[i]) / norm)
		}
	}
	return vec, nil
}

const defaultOpenAIEmbeddingsURL = "https://api.openai.com/v1/embeddings"

// OpenAIEmbedder produces real semantic embeddings via the OpenAI embeddings
// API.
//
// It sends the `dimensions` parameter to truncate text-embedding-3-* output to
// dim, so it can fill the existing pgvector column (kb_chunks.embedding is a
// fixed vector(EMBEDDING_DIM)) without a schema migration.
type OpenAIEmbedder struct {
	apiKey  string
	model   string
	dim     int
	baseURL string
	client  *http.Client
}

// OpenAIEmbedderOption customises an OpenAIEmbedder (used by tests to point at
// a stub server or inject a client).
type OpenAIEmbedderOption func(*OpenAIEmbedder)

// WithOpenAIBaseURL overrides the embeddings endpoint.
func WithOpenAIBaseURL(url string) OpenAIEmbedderOption {
	return func(e *OpenAIEmbedder) { e.baseURL = url }
}

// WithOpenAIHTTPClient overrides the HTTP client.
func WithOpenAIHTTPClient(c *http.Client) OpenAIEmbedderOption {
	return func(e *OpenAIEmbedder) { e.client = c }
}

// NewOpenAIEmbedder builds an embedder backed by the OpenAI embeddings API.
func NewOpenAIEmbedder(apiKey, model string, dim int, opts ...OpenAIEmbedderOption) *OpenAIEmbedder {
	e := &OpenAIEmbedder{
		apiKey:  apiKey,
		model:   model,
		dim:     dim,
		baseURL: defaultOpenAIEmbeddingsURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *OpenAIEmbedder) Dim() int { return e.dim }

func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	body, err := json.Marshal(map[string]any{
		"model":      e.model,
		"input":      text,
		"dimensions": e.dim,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+e.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai embeddings: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai embeddings: unexpected status %d", resp.StatusCode)
	}

	var payload struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("openai embeddings: decode: %w", err)
	}
	if len(payload.Data) == 0 {
		return nil, fmt.Errorf("openai embeddings: empty data")
	}
	return payload.Data[0].Embedding, nil
}

// EmbedderSettings is the slice of configuration BuildEmbedder needs, so this
// package doesn't depend on the config package.
type EmbedderSettings struct {
	OpenAIAPIKey   string
	EmbeddingModel string
	EmbeddingDim   int
}

// BuildEmbedder picks the configured embedder: OpenAI when an API key is set,
// otherwise the offline hashing embedder (tests / no-key local runs).
func BuildEmbedder(s EmbedderSettings) Embedder {
	if s.OpenAIAPIKey != "" {
		return NewOpenAIEmbedder(s.OpenAIAPIKey, s.EmbeddingModel, s.EmbeddingDim)
	}
	return NewHashingEmbedder(s.EmbeddingDim)
}
