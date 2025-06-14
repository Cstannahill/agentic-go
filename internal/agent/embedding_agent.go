package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"agentic.example.com/mvp/internal/tools"
	"agentic.example.com/mvp/internal/vectorstore"
)

// EmbeddingAgent generates embeddings and stores them in the vector store.
type EmbeddingAgent struct {
	id    string
	store vectorstore.VectorStore
	tool  *tools.EmbeddingTool
}

// NewEmbeddingAgent returns a new EmbeddingAgent using the default store.
func NewEmbeddingAgent() *EmbeddingAgent {
	return &EmbeddingAgent{
		id:    fmt.Sprintf("embedding-agent-%s", uuid.NewString()),
		store: vectorstore.DefaultStore(),
		tool:  tools.NewEmbeddingToolWithProvider(tools.DefaultEmbeddingProvider()),
	}
}

func (e *EmbeddingAgent) ID() string { return e.id }

func (e *EmbeddingAgent) Execute(ctx context.Context, task Task) Result {
	out, err := e.tool.Run(ctx, task.Input)
	if err != nil {
		return Result{TaskID: task.ID, Error: err}
	}
	emb := out["embedding"].([]float64)
	if e.store != nil {
		doc := vectorstore.Document{ID: task.ID, Embedding: emb, Metadata: task.Input}
		e.store.Upsert(ctx, []vectorstore.Document{doc})
	}
	return Result{TaskID: task.ID, Output: out, Successful: true}
}

func init() {
	Register("EmbeddingAgent", func() Agent { return NewEmbeddingAgent() })
}
