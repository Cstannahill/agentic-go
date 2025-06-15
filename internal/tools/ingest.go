package tools

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"agentic.example.com/mvp/internal/vectorstore"
)

// IngestTool embeds text using an EmbeddingProvider and stores the resulting
// vector in a VectorStore. It returns the document ID that was persisted.
//
// Expected input fields:
//
//	text     - string of content to embed
//	id       - optional unique identifier for the document. If not provided a
//	           uuid will be generated.
//	metadata - optional map of additional payload values to store alongside the
//	           vector.
type IngestTool struct {
	Store    vectorstore.VectorStore
	Embedder *EmbeddingTool
}

// NewIngestTool constructs an IngestTool using default provider and store.
func NewIngestTool() *IngestTool {
	return &IngestTool{Embedder: &EmbeddingTool{}}
}

// Run implements the Tool interface.
func (i *IngestTool) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	text, _ := input["text"].(string)
	if text == "" {
		return nil, errors.New("text field required")
	}
	id, _ := input["id"].(string)
	metadata, _ := input["metadata"].(map[string]interface{})

	if i.Embedder == nil {
		i.Embedder = &EmbeddingTool{}
	}
	if i.Store == nil {
		i.Store = vectorstore.DefaultStore()
	}

	out, err := i.Embedder.Run(ctx, map[string]interface{}{"text": text})
	if err != nil {
		return nil, err
	}
	emb := out["embedding"].([]float64)

	doc := vectorstore.Document{ID: id, Embedding: emb, Metadata: metadata}
	if doc.ID == "" {
		doc.ID = generateID()
	}
	if err := i.Store.Upsert(ctx, []vectorstore.Document{doc}); err != nil {
		return nil, err
	}
	return map[string]interface{}{"id": doc.ID}, nil
}

// generateID returns a uuid string. Placed in a helper for testability.
func generateID() string { return uuidNewString() }

// uuidNewString is defined as a variable for patching in tests.
var uuidNewString = func() string { return uuid.NewString() }
