package memai

import "context"

// MemoryStore is the interface for long-term memory persistence.
// Implementations can use SQLite, PostgreSQL, or any other backend.
type MemoryStore interface {
	// GetMemories returns all memories with embeddings for a user.
	GetMemories(ctx context.Context, userID string) ([]Memory, error)

	// SaveMemory persists a new memory.
	SaveMemory(ctx context.Context, mem *Memory) error

	// DeleteMemory removes a memory by ID.
	DeleteMemory(ctx context.Context, id int64) error

	// UpdateBoost adjusts the feedback boost of a memory.
	UpdateBoost(ctx context.Context, id int64, delta float64) error
}

// EmbeddingFunc generates an embedding vector for the given text.
// This decouples the memory system from any specific embedding provider.
type EmbeddingFunc func(ctx context.Context, text string) ([]float64, error)
