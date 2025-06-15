package tools

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"agentic.example.com/mvp/internal/vectorstore"
)

func TestIngestTool(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	vectorstore.SetDefaultStore(store)

	SetDefaultEmbeddingProvider(HashEmbeddingProvider{Dim: 8})

	fixedID := "doc1"
	uuidNewString = func() string { return fixedID }
	defer func() { uuidNewString = uuid.NewString }()

	tool := NewIngestTool()
	out, err := tool.Run(context.Background(), map[string]interface{}{"text": "hello"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if out["id"] != fixedID {
		t.Fatalf("unexpected id %v", out["id"])
	}

        docs, err := store.Query(context.Background(), vectorstore.QueryRequest{Embedding: BasicHashEmbed("hello", 8), TopK: 1})
	if err != nil || len(docs) == 0 || docs[0].ID != fixedID {
		t.Fatalf("document not stored: %v %v", err, docs)
	}
}
