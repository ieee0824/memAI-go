package memai

// EmotionType represents a classified emotion.
type EmotionType string

const (
	EmotionJoy      EmotionType = "joy"
	EmotionSadness  EmotionType = "sadness"
	EmotionAnger    EmotionType = "anger"
	EmotionFear     EmotionType = "fear"
	EmotionSurprise EmotionType = "surprise"
	EmotionNeutral  EmotionType = "neutral"
)

// EmotionalState represents the detected emotional context of a message.
type EmotionalState struct {
	Primary   EmotionType
	Intensity float64 // 0.0 - 1.0
	Valence   float64 // -1.0 to 1.0
}

// WorkingMemoryItem is an active item in short-term memory.
type WorkingMemoryItem struct {
	Topic        string
	Content      string
	Keywords     []string
	Activation   float64 // 0.0 - 1.0
	TurnCreated  int
	TurnAccessed int
	Emotional    bool
}

// Memory represents a stored long-term memory with its embedding.
type Memory struct {
	ID                 int64
	Content            string
	Embedding          []float64
	ThreadKey          string
	EventDate          string
	Boost              float64
	EmotionalIntensity float64
}

// SearchResult represents a memory search result with computed score.
type SearchResult struct {
	Memory Memory
	Score  float64
}

// SearchQuery holds the parameters for a long-term memory search.
type SearchQuery struct {
	Query              string
	QueryEmbedding     []float64
	ThreadKey          string
	QueryDate          string
	DateNegated        bool
	DateMonthOnly      bool
	EmotionalIntensity float64
}
