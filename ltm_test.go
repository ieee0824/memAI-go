package memai

import (
	"context"
	"math"
	"testing"
)

// mockStore implements MemoryStore[int] for testing.
type mockStore struct {
	memories []Memory[int]
}

func (m *mockStore) GetMemories(_ context.Context) ([]Memory[int], error) {
	return m.memories, nil
}
func (m *mockStore) SaveMemory(_ context.Context, mem *Memory[int]) error {
	m.memories = append(m.memories, *mem)
	return nil
}
func (m *mockStore) DeleteMemory(_ context.Context, id int) error { return nil }
func (m *mockStore) UpdateBoost(_ context.Context, id int, delta float64) error {
	for i := range m.memories {
		if m.memories[i].ID == id {
			m.memories[i].Boost += delta
		}
	}
	return nil
}

func TestCosineSimilarity_Identical(t *testing.T) {
	v := []float64{1, 0, 0}
	sim := CosineSimilarity(v, v)
	if math.Abs(sim-1.0) > 0.001 {
		t.Errorf("expected 1.0, got %f", sim)
	}
}

func TestCosineSimilarity_Orthogonal(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{0, 1, 0}
	sim := CosineSimilarity(a, b)
	if math.Abs(sim) > 0.001 {
		t.Errorf("expected 0, got %f", sim)
	}
}

func TestCosineSimilarity_DifferentLength(t *testing.T) {
	a := []float64{1, 0}
	b := []float64{1, 0, 0}
	sim := CosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("expected 0 for different lengths, got %f", sim)
	}
}

func TestLTM_Search(t *testing.T) {
	store := &mockStore{
		memories: []Memory[int]{
			{ID: 1, Content: "relevant", Embedding: []float64{0.9, 0.1, 0.0}},
			{ID: 2, Content: "irrelevant", Embedding: []float64{0.0, 0.0, 1.0}},
		},
	}
	ltm := NewLTM(store, nil, DefaultLTMConfig())

	results, err := ltm.Search(context.Background(), SearchQuery{
		QueryEmbedding: []float64{1.0, 0.0, 0.0},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if results[0].Memory.Content != "relevant" {
		t.Errorf("expected 'relevant' first, got %q", results[0].Memory.Content)
	}
}

func TestLTM_EmotionalPriming(t *testing.T) {
	store := &mockStore{
		memories: []Memory[int]{
			{ID: 1, Content: "borderline", Embedding: []float64{0.6, 0.5, 0.0}},
		},
	}
	ltm := NewLTM(store, nil, DefaultLTMConfig())

	emb := []float64{1.0, 0.0, 0.0}

	// Without emotional priming
	r1, _ := ltm.Search(context.Background(), SearchQuery{
		QueryEmbedding:     emb,
		EmotionalIntensity: 0.0,
	})

	// With emotional priming (lower threshold)
	r2, _ := ltm.Search(context.Background(), SearchQuery{
		QueryEmbedding:     emb,
		EmotionalIntensity: 0.8,
	})

	if len(r2) < len(r1) {
		t.Error("emotional priming should return equal or more results")
	}
}

func TestLTM_ThreadBoost(t *testing.T) {
	store := &mockStore{
		memories: []Memory[int]{
			{ID: 1, Content: "same-thread", Embedding: []float64{0.8, 0.2, 0.0}, ThreadKey: "t1"},
			{ID: 2, Content: "other-thread", Embedding: []float64{0.8, 0.2, 0.0}, ThreadKey: "t2"},
		},
	}
	ltm := NewLTM(store, nil, DefaultLTMConfig())

	results, _ := ltm.Search(context.Background(), SearchQuery{
		QueryEmbedding: []float64{1.0, 0.0, 0.0},
		ThreadKey:      "t1",
	})

	if len(results) < 2 {
		t.Fatal("expected 2 results")
	}
	if results[0].Memory.Content != "same-thread" {
		t.Error("same-thread memory should rank higher")
	}
}

// Regression (#4): a large feedback Boost must not admit a semantically
// irrelevant (orthogonal) memory below the cosine threshold.
func TestLTM_BoostDoesNotBypassThreshold(t *testing.T) {
	store := &mockStore{
		memories: []Memory[int]{
			{ID: 1, Content: "irrelevant", Embedding: []float64{0, 1, 0}, Boost: 5.0},
		},
	}
	ltm := NewLTM(store, nil, DefaultLTMConfig())
	results, err := ltm.Search(context.Background(), SearchQuery{QueryEmbedding: []float64{1, 0, 0}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("orthogonal memory with high boost should be excluded, got %d results", len(results))
	}
}

// Regression (#5): TopK <= 0 means "no limit", not "return nothing".
func TestLTM_TopKZeroMeansNoLimit(t *testing.T) {
	cfg := DefaultLTMConfig()
	cfg.TopK = 0
	store := &mockStore{
		memories: []Memory[int]{
			{ID: 1, Content: "a", Embedding: []float64{1, 0, 0}},
			{ID: 2, Content: "b", Embedding: []float64{0.9, 0.1, 0}},
		},
	}
	ltm := NewLTM(store, nil, cfg)
	results, err := ltm.Search(context.Background(), SearchQuery{QueryEmbedding: []float64{1, 0, 0}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("TopK=0 should not truncate; expected 2 results, got %d", len(results))
	}
}

// Regression (#6): date matching normalizes formatting differences and reports
// ok=false for unparseable input.
func TestDateMatches_Normalization(t *testing.T) {
	cases := []struct {
		q, m        string
		monthOnly   bool
		wantMatch   bool
		wantParseOK bool
	}{
		{"2026-6-17", "2026-06-17", false, true, true},            // non-padded vs padded
		{"2026/06/17", "2026-06-17", false, true, true},           // slash vs dash
		{"2026-06-17T10:00:00Z", "2026-06-17", false, true, true}, // RFC3339 vs date
		{"2026-06-18", "2026-06-17", false, false, true},          // different day
		{"2026-06-01", "2026-06-30", true, true, true},            // month-only match
		{"not-a-date", "2026-06-17", false, false, false},
	}
	for _, c := range cases {
		matched, ok := dateMatches(c.q, c.m, c.monthOnly)
		if ok != c.wantParseOK {
			t.Errorf("dateMatches(%q,%q,%v) ok=%v, want %v", c.q, c.m, c.monthOnly, ok, c.wantParseOK)
		}
		if ok && matched != c.wantMatch {
			t.Errorf("dateMatches(%q,%q,%v) matched=%v, want %v", c.q, c.m, c.monthOnly, matched, c.wantMatch)
		}
	}
}
