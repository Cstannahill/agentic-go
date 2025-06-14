package vectorstore

import "context"

// Document represents an item stored in the vector store.
type Document struct {
	ID        string
	Embedding []float64
	Metadata  map[string]interface{}
	Score     float64 // optional score returned by queries
}

// VectorStore defines the operations supported by a store.
type VectorStore interface {
	Upsert(ctx context.Context, docs []Document) error
	Query(ctx context.Context, embedding []float64, k int) ([]Document, error)
	Delete(ctx context.Context, ids []string) error
}

var defaultStore VectorStore

// SetDefaultStore configures the global store used by default.
func SetDefaultStore(store VectorStore) { defaultStore = store }

// DefaultStore returns the globally configured store.
func DefaultStore() VectorStore { return defaultStore }
