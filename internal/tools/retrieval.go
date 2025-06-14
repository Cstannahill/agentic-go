package tools

import (
	"context"
	"errors"

	"agentic.example.com/mvp/internal/vectorstore"
)

// RetrievalTool retrieves nearest documents from a vector store.
type RetrievalTool struct {
	Store vectorstore.VectorStore
	TopK  int
}

// NewRetrievalTool constructs a RetrievalTool.
func NewRetrievalTool(store vectorstore.VectorStore, k int) *RetrievalTool {
	if k <= 0 {
		k = DefaultTopK()
	}
	return &RetrievalTool{Store: store, TopK: k}
}

var defaultTopK = 5

// SetDefaultTopK sets the global default for retrieval when none is provided.
func SetDefaultTopK(k int) {
	if k > 0 {
		defaultTopK = k
	}
}

// DefaultTopK returns the currently configured default top K.
func DefaultTopK() int { return defaultTopK }

// Run implements the Tool interface.
func (r *RetrievalTool) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	emb, ok := input["embedding"].([]float64)
	if !ok || len(emb) == 0 {
		return nil, errors.New("embedding required")
	}
	if r.Store == nil {
		r.Store = vectorstore.DefaultStore()
	}
	docs, err := r.Store.Query(ctx, emb, r.TopK)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, len(docs))
	for i, d := range docs {
		out[i] = map[string]interface{}{
			"id":       d.ID,
			"metadata": d.Metadata,
			"score":    d.Score,
		}
	}
	return map[string]interface{}{"documents": out}, nil
}
