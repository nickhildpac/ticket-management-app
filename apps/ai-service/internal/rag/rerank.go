package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

// DefaultOpenRouterRerankURL is OpenRouter's Cohere-style re-rank endpoint.
const DefaultOpenRouterRerankURL = "https://openrouter.ai/api/v1/rerank"

// Reranker scores (query, passage) pairs; higher is more relevant.
type Reranker interface {
	Score(ctx context.Context, query string, passages []string) ([]float64, error)
}

// OpenRouterReranker performs a Cohere-style re-rank via OpenRouter's
// /api/v1/rerank endpoint.
//
// The default model is cohere/rerank-v3.5. Scores are returned aligned to the
// input passage order (the API returns ranked results; we map index -> score).
type OpenRouterReranker struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// RerankerOption customises an OpenRouterReranker.
type RerankerOption func(*OpenRouterReranker)

// WithRerankBaseURL overrides the re-rank endpoint.
func WithRerankBaseURL(url string) RerankerOption {
	return func(r *OpenRouterReranker) { r.baseURL = strings.TrimRight(url, "/") }
}

// WithRerankHTTPClient overrides the HTTP client.
func WithRerankHTTPClient(c *http.Client) RerankerOption {
	return func(r *OpenRouterReranker) { r.client = c }
}

// NewOpenRouterReranker builds a cross-encoder re-ranker backed by OpenRouter.
func NewOpenRouterReranker(apiKey, model string, opts ...RerankerOption) *OpenRouterReranker {
	r := &OpenRouterReranker{
		apiKey:  apiKey,
		model:   model,
		baseURL: DefaultOpenRouterRerankURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *OpenRouterReranker) Score(ctx context.Context, query string, passages []string) ([]float64, error) {
	if len(passages) == 0 {
		return nil, nil
	}
	if strings.TrimSpace(r.apiKey) == "" {
		return nil, errors.New(
			"OPENROUTER_API_KEY is empty — cross-encoder re-rank requires it " +
				"(model=cohere/rerank-v3.5 via OpenRouter)")
	}

	body, err := json.Marshal(map[string]any{
		"model":     r.model,
		"query":     query,
		"documents": passages,
		"top_n":     len(passages),
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter rerank: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openrouter rerank: unexpected status %d", resp.StatusCode)
	}

	var payload struct {
		Results []struct {
			Index          int     `json:"index"`
			RelevanceScore float64 `json:"relevance_score"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("openrouter rerank: decode: %w", err)
	}

	// The API returns ranked results carrying the original document indices;
	// fill a dense score vector so callers can zip against the input passages.
	scores := make([]float64, len(passages))
	for _, item := range payload.Results {
		if item.Index >= 0 && item.Index < len(passages) {
			scores[item.Index] = item.RelevanceScore
		}
	}
	return scores, nil
}

// ScoreToDistance maps a higher-is-better re-ranker score to a lower-is-better
// distance. Works for both logits and [0, 1] relevance scores (it is monotone
// decreasing either way).
func ScoreToDistance(score float64) float64 {
	prob := 1.0 / (1.0 + math.Exp(-score))
	return 1.0 - prob
}
