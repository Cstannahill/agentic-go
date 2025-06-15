package setup

import (
	"os"
	"testing"

	"agentic.example.com/mvp/internal/vectorstore"
)

func TestInitFromEnv(t *testing.T) {
	os.Setenv("EMBEDDING_DIM", "32")
	os.Setenv("VECTORSTORE_ENDPOINT", "")
	t.Cleanup(func() {
		os.Unsetenv("EMBEDDING_DIM")
		os.Unsetenv("VECTORSTORE_ENDPOINT")
	})

	cfg := InitFromEnv()
	if cfg.EmbeddingDim != 32 {
		t.Fatalf("expected embedding dim 32, got %d", cfg.EmbeddingDim)
	}
	if vectorstore.DefaultStore() == nil {
		t.Fatalf("default store not initialised")
	}
}
