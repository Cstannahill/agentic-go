package vectorstore

import "context"

// Document represents an item stored in the vector store.
type Document struct {
	ID        string
	Embedding []float64
	Metadata  map[string]interface{}
	Score     float64 // optional score returned by queries
}

// QueryRequest describes a retrieval operation.
type QueryRequest struct {
	Embedding []float64
	TopK      int
	Filter    map[string]interface{}
}

// VectorStore defines the operations supported by a store.
type VectorStore interface {
	Upsert(ctx context.Context, docs []Document) error
	Query(ctx context.Context, req QueryRequest) ([]Document, error)
	Delete(ctx context.Context, ids []string) error
}

var defaultStore VectorStore

// SetDefaultStore configures the global store used by default.
func SetDefaultStore(store VectorStore) { defaultStore = store }

// DefaultStore returns the globally configured store.
func DefaultStore() VectorStore { return defaultStore }
