package memai

import "context"

// MemoryStore is the interface for long-term memory persistence.
// Implementations can use SQLite, PostgreSQL, or any other backend.
type MemoryStore[ID comparable] interface {
	// GetMemories returns all memories with embeddings.
	GetMemories(ctx context.Context) ([]Memory[ID], error)

	// SaveMemory persists a new memory.
	SaveMemory(ctx context.Context, mem *Memory[ID]) error

	// DeleteMemory removes a memory by ID.
	DeleteMemory(ctx context.Context, id ID) error

	// UpdateBoost adjusts the feedback boost of a memory.
	UpdateBoost(ctx context.Context, id ID, delta float64) error
}

// EmbeddingFunc generates an embedding vector for the given text.
// This decouples the memory system from any specific embedding provider.
type EmbeddingFunc func(ctx context.Context, text string) ([]float64, error)
