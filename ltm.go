package memai

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
)

// LTMConfig configures long-term memory search behavior.
type LTMConfig struct {
	SimilarityThreshold float64 // Minimum cosine similarity to include (default: 0.3)
	TopK                int     // Maximum results to return (default: 10)
	ThreadBoost         float64 // Score boost for same-thread memories (default: 0.1)
	DateBoost           float64 // Score boost for matching date (default: 0.15)
	DatePenalty         float64 // Score penalty for mismatched date (default: -0.2)
	EmotionalBoost      float64 // Score boost factor for emotional memories (default: 0.12)
	EmotionalPrimeDelta float64 // Threshold reduction when user is emotional (default: 0.05)
}

// DefaultLTMConfig returns the default long-term memory configuration.
func DefaultLTMConfig() LTMConfig {
	return LTMConfig{
		SimilarityThreshold: 0.3,
		TopK:                10,
		ThreadBoost:         0.1,
		DateBoost:           0.15,
		DatePenalty:         -0.2,
		EmotionalBoost:      0.12,
		EmotionalPrimeDelta: 0.05,
	}
}

// LTM manages long-term memory search with emotional priming.
type LTM struct {
	config    LTMConfig
	store     MemoryStore
	embedding EmbeddingFunc
}

// NewLTM creates a new long-term memory manager.
func NewLTM(store MemoryStore, embeddingFn EmbeddingFunc, config LTMConfig) *LTM {
	return &LTM{
		config:    config,
		store:     store,
		embedding: embeddingFn,
	}
}

// Search finds relevant memories for the given query using vector similarity
// with multi-factor scoring (thread, date, emotion) and emotional priming.
func (l *LTM) Search(ctx context.Context, q SearchQuery) ([]SearchResult, error) {
	// Generate embedding if not provided
	queryEmb := q.QueryEmbedding
	if len(queryEmb) == 0 {
		if l.embedding == nil {
			return nil, fmt.Errorf("no embedding function and no query embedding provided")
		}
		var err error
		queryEmb, err = l.embedding(ctx, q.Query)
		if err != nil {
			return nil, fmt.Errorf("embedding generation failed: %w", err)
		}
	}

	memories, err := l.store.GetMemories(ctx, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("memory store error: %w", err)
	}

	// Emotional priming: lower threshold when user is emotional
	threshold := l.config.SimilarityThreshold
	if q.EmotionalIntensity > 0.5 {
		threshold -= l.config.EmotionalPrimeDelta
	}

	var results []SearchResult
	for _, mem := range memories {
		if len(mem.Embedding) == 0 {
			continue
		}

		score := CosineSimilarity(queryEmb, mem.Embedding)

		// Feedback boost
		score += mem.Boost

		// Emotional boost
		score += l.config.EmotionalBoost * mem.EmotionalIntensity

		// Thread boost
		if q.ThreadKey != "" && mem.ThreadKey == q.ThreadKey {
			score += l.config.ThreadBoost
		}

		// Date boost/penalty
		if q.QueryDate != "" && mem.EventDate != "" {
			if dateMatches(q.QueryDate, mem.EventDate, q.DateMonthOnly) {
				if q.DateNegated {
					score += l.config.DatePenalty
				} else {
					score += l.config.DateBoost
				}
			} else {
				if q.DateNegated {
					score += l.config.DateBoost
				} else {
					score += l.config.DatePenalty
				}
			}
		}

		if score >= threshold {
			results = append(results, SearchResult{Memory: mem, Score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > l.config.TopK {
		results = results[:l.config.TopK]
	}

	return results, nil
}

// ApplyFeedback adjusts the boost value for the given memories.
func (l *LTM) ApplyFeedback(ctx context.Context, memoryIDs []int64, delta float64) error {
	for _, id := range memoryIDs {
		if err := l.store.UpdateBoost(ctx, id, delta); err != nil {
			return err
		}
	}
	return nil
}

// FormatResults formats search results as a human-readable context string.
func FormatResults(results []SearchResult) string {
	if len(results) == 0 {
		return ""
	}

	var lines []string
	for _, r := range results {
		lines = append(lines, fmt.Sprintf("- %s", r.Memory.Content))
	}
	return strings.Join(lines, "\n")
}

// dateMatches checks whether two date strings match.
// If monthOnly is true, only compares year-month (YYYY-MM).
func dateMatches(queryDate, memDate string, monthOnly bool) bool {
	if monthOnly {
		// Compare YYYY-MM prefix
		if len(queryDate) >= 7 && len(memDate) >= 7 {
			return queryDate[:7] == memDate[:7]
		}
		return false
	}
	// Compare full YYYY-MM-DD (first 10 chars)
	if len(queryDate) >= 10 && len(memDate) >= 10 {
		return queryDate[:10] == memDate[:10]
	}
	return queryDate == memDate
}

// CosineSimilarity computes the cosine similarity between two vectors.
// Returns 0 if either vector is zero-length or they have different dimensions.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return dot / denom
}
