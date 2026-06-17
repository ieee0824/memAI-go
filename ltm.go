package memai

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// LTMConfig configures long-term memory search behavior.
//
// Inclusion is gated on cosine similarity alone (>= SimilarityThreshold, lowered
// by EmotionalPrimeDelta when the user is emotional). The remaining factors
// (feedback Boost, emotion, thread, date) only adjust the score used for
// ranking the included results; they never resurrect a semantically irrelevant
// memory.
type LTMConfig struct {
	SimilarityThreshold float64 // Minimum cosine similarity to include (default: 0.3)
	TopK                int     // Maximum results to return; <= 0 means no limit (default: 10)
	ThreadBoost         float64 // Ranking boost for same-thread memories (default: 0.1)
	DateBoost           float64 // Ranking boost for matching date (default: 0.15)
	DatePenalty         float64 // Ranking penalty for mismatched date (default: -0.2)
	EmotionalBoost      float64 // Ranking boost factor for emotional memories (default: 0.12)
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
type LTM[ID comparable] struct {
	config    LTMConfig
	store     MemoryStore[ID]
	embedding EmbeddingFunc
}

// NewLTM creates a new long-term memory manager.
func NewLTM[ID comparable](store MemoryStore[ID], embeddingFn EmbeddingFunc, config LTMConfig) *LTM[ID] {
	return &LTM[ID]{
		config:    config,
		store:     store,
		embedding: embeddingFn,
	}
}

// Search finds relevant memories for the given query using vector similarity
// with multi-factor scoring (thread, date, emotion) and emotional priming.
func (l *LTM[ID]) Search(ctx context.Context, q SearchQuery) ([]SearchResult[ID], error) {
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

	memories, err := l.store.GetMemories(ctx)
	if err != nil {
		return nil, fmt.Errorf("memory store error: %w", err)
	}

	// Emotional priming: lower threshold when user is emotional
	threshold := l.config.SimilarityThreshold
	if q.EmotionalIntensity > 0.5 {
		threshold -= l.config.EmotionalPrimeDelta
	}

	var results []SearchResult[ID]
	for _, mem := range memories {
		if len(mem.Embedding) == 0 {
			continue
		}

		// Inclusion is decided by cosine similarity alone (with emotional
		// priming); the boosts below only affect ranking.
		sim := CosineSimilarity(queryEmb, mem.Embedding)
		if sim < threshold {
			continue
		}

		score := sim

		// Feedback boost
		score += mem.Boost

		// Emotional boost
		score += l.config.EmotionalBoost * mem.EmotionalIntensity

		// Thread boost
		if q.ThreadKey != "" && mem.ThreadKey == q.ThreadKey {
			score += l.config.ThreadBoost
		}

		// Date boost/penalty (ranking only)
		score += l.dateDelta(q, mem)

		results = append(results, SearchResult[ID]{Memory: mem, Score: score})
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if l.config.TopK > 0 && len(results) > l.config.TopK {
		results = results[:l.config.TopK]
	}

	return results, nil
}

// dateDelta returns the ranking adjustment for the date factor. It is zero
// when either date is empty or cannot be parsed, so a malformed or
// foreign-format date never penalizes a memory.
func (l *LTM[ID]) dateDelta(q SearchQuery, mem Memory[ID]) float64 {
	if q.QueryDate == "" || mem.EventDate == "" {
		return 0
	}
	matched, ok := dateMatches(q.QueryDate, mem.EventDate, q.DateMonthOnly)
	if !ok {
		return 0
	}
	// matched != negated => the date is "as wanted" => boost; otherwise penalty.
	if matched != q.DateNegated {
		return l.config.DateBoost
	}
	return l.config.DatePenalty
}

// ApplyFeedback adjusts the boost value for the given memories.
func (l *LTM[ID]) ApplyFeedback(ctx context.Context, memoryIDs []ID, delta float64) error {
	for _, id := range memoryIDs {
		if err := l.store.UpdateBoost(ctx, id, delta); err != nil {
			return err
		}
	}
	return nil
}

// dateLayouts are the accepted date formats, tried in order. Both zero-padded
// and non-padded numeric forms are accepted, along with RFC3339 timestamps and
// year-month-only values.
var dateLayouts = []string{
	"2006-01-02", "2006-1-2", "2006/01/02", "2006/1/2",
	time.RFC3339, "2006-01-02T15:04:05",
	"2006-01", "2006-1", "2006/01", "2006/1",
}

// parseDate parses a date string using dateLayouts, returning ok=false if no
// layout matches.
func parseDate(s string) (time.Time, bool) {
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// dateMatches reports whether two date strings refer to the same date (or, when
// monthOnly is true, the same year-month). ok is false when either string
// cannot be parsed, so the caller can skip the date factor rather than treating
// a parse failure as a mismatch. Parsing normalizes formatting differences
// (e.g. "2026-6-17" matches "2026-06-17").
func dateMatches(queryDate, memDate string, monthOnly bool) (matched, ok bool) {
	q, ok1 := parseDate(queryDate)
	m, ok2 := parseDate(memDate)
	if !ok1 || !ok2 {
		return false, false
	}
	if q.Year() != m.Year() || q.Month() != m.Month() {
		return false, true
	}
	if monthOnly {
		return true, true
	}
	return q.Day() == m.Day(), true
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
