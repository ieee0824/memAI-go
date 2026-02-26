package memai

import (
	"context"
	"math"
	"testing"
)

// mockStore implements MemoryStore for testing.
type mockStore struct {
	memories []Memory
}

func (m *mockStore) GetMemories(_ context.Context, _ string) ([]Memory, error) {
	return m.memories, nil
}
func (m *mockStore) SaveMemory(_ context.Context, mem *Memory) error {
	m.memories = append(m.memories, *mem)
	return nil
}
func (m *mockStore) DeleteMemory(_ context.Context, id int64) error { return nil }
func (m *mockStore) UpdateBoost(_ context.Context, id int64, delta float64) error {
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
		memories: []Memory{
			{ID: 1, Content: "relevant", Embedding: []float64{0.9, 0.1, 0.0}},
			{ID: 2, Content: "irrelevant", Embedding: []float64{0.0, 0.0, 1.0}},
		},
	}
	ltm := NewLTM(store, nil, DefaultLTMConfig())

	results, err := ltm.Search(context.Background(), SearchQuery{
		UserID:         "u1",
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
		memories: []Memory{
			{ID: 1, Content: "borderline", Embedding: []float64{0.6, 0.5, 0.0}},
		},
	}
	ltm := NewLTM(store, nil, DefaultLTMConfig())

	emb := []float64{1.0, 0.0, 0.0}

	// Without emotional priming
	r1, _ := ltm.Search(context.Background(), SearchQuery{
		UserID:             "u1",
		QueryEmbedding:     emb,
		EmotionalIntensity: 0.0,
	})

	// With emotional priming (lower threshold)
	r2, _ := ltm.Search(context.Background(), SearchQuery{
		UserID:             "u1",
		QueryEmbedding:     emb,
		EmotionalIntensity: 0.8,
	})

	if len(r2) < len(r1) {
		t.Error("emotional priming should return equal or more results")
	}
}

func TestLTM_ThreadBoost(t *testing.T) {
	store := &mockStore{
		memories: []Memory{
			{ID: 1, Content: "same-thread", Embedding: []float64{0.8, 0.2, 0.0}, ThreadKey: "t1"},
			{ID: 2, Content: "other-thread", Embedding: []float64{0.8, 0.2, 0.0}, ThreadKey: "t2"},
		},
	}
	ltm := NewLTM(store, nil, DefaultLTMConfig())

	results, _ := ltm.Search(context.Background(), SearchQuery{
		UserID:         "u1",
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

func TestFormatResults(t *testing.T) {
	results := []SearchResult{
		{Memory: Memory{Content: "fact 1"}, Score: 0.9},
		{Memory: Memory{Content: "fact 2"}, Score: 0.8},
	}

	formatted := FormatResults(results)
	if formatted == "" {
		t.Error("expected non-empty format")
	}
}

func TestFormatResults_Empty(t *testing.T) {
	if FormatResults(nil) != "" {
		t.Error("expected empty for nil results")
	}
}
