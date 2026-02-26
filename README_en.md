# memAI-go

A Go library for brain-inspired memory systems. Provides amygdala-based emotion detection, short-term memory (working memory), and long-term memory (vector search + emotional priming).

## Installation

```bash
go get github.com/ieee0824/memAI-go
```

## Package Structure

```
package memai

├── emotion.go   # Emotion detection (amygdala)
├── stm.go       # Short-term memory (working memory)
├── ltm.go       # Long-term memory (vector search)
├── feedback.go  # Feedback detection
├── store.go     # Storage interface
└── types.go     # Common type definitions
```

## Usage

### Emotion Detection

Keyword-based Japanese emotion detection. No LLM required.

```go
es := memai.AnalyzeEmotion("嬉しい！ありがとう！")
// es.Primary   = EmotionJoy
// es.Intensity = 0.7  (0.0-1.0)
// es.Valence   = 0.56 (-1.0 to 1.0)
```

6 categories: joy, sadness, anger, fear, surprise, neutral

### Short-Term Memory (STM)

Activation-decay model for working memory. Emotional items decay more slowly.

```go
stm := memai.NewSTM(memai.DefaultSTMConfig())

// Add an item
stm.Add(&memai.WorkingMemoryItem{
    Topic:        "meeting schedule",
    Keywords:     []string{"meeting", "schedule"},
    Activation:   1.0,
    TurnCreated:  0,
    TurnAccessed: 0,
})

// Turn update: decay → emotional marking → keyword refresh → eviction
emotion := memai.AnalyzeEmotion(userMessage)
stm.Update(currentTurn, userMessage, emotion)

// Formatted output
fmt.Println(stm.Format())
// - [高] meeting schedule
```

Configuration:
- `MaxItems`: Working memory capacity (default: 7)
- `NormalDecayRate`: Activation decay per turn (default: 0.15/turn)
- `EmotionalDecayRate`: Decay rate for emotional items (default: 0.07/turn)
- `RefreshBoost`: Activation boost on keyword match (default: +0.3)
- `ActivationThreshold`: Eviction threshold (default: 0.1)

### Long-Term Memory (LTM)

Vector search + multi-factor scoring + emotional priming.

```go
ltm := memai.NewLTM(store, embeddingFn, memai.DefaultLTMConfig())

results, err := ltm.Search(ctx, memai.SearchQuery{
    Query:              "tomorrow's meeting",
    ThreadKey:          "thread-1",
    EmotionalIntensity: 0.7,
})

// Apply feedback
ltm.ApplyFeedback(ctx, memoryIDs, memai.FeedbackBoostPositive)
```

Scoring factors:
- **Cosine similarity**: Similarity between embeddings
- **Emotional boost**: Prioritizes emotional memories (+0.12 x intensity)
- **Thread boost**: Prioritizes same-thread memories (+0.1)
- **Date boost/penalty**: Date match (+0.15) / mismatch (-0.2)
- **Emotional priming**: Lowers threshold when user is emotional (0.3 → 0.25)

### Feedback Detection

Detects memory accuracy feedback from user responses.

```go
delta := memai.DetectFeedback("ありがとう、そうそう！")
// delta = +0.05 (positive)

delta = memai.DetectFeedback("違うよ、それじゃない")
// delta = -0.05 (negative)
```

### Storage Interface

Implement `MemoryStore` to use any backend.

```go
type MemoryStore interface {
    GetMemories(ctx context.Context) ([]Memory, error)
    SaveMemory(ctx context.Context, mem *Memory) error
    DeleteMemory(ctx context.Context, id int64) error
    UpdateBoost(ctx context.Context, id int64, delta float64) error
}
```

`EmbeddingFunc` allows swapping the embedding provider.

```go
type EmbeddingFunc func(ctx context.Context, text string) ([]float64, error)
```

## License

MIT
