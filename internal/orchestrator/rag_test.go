package orchestrator

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"agentic.example.com/mvp/internal/tools"
	"agentic.example.com/mvp/internal/vectorstore"
)

// TestResolveSourcePath ensures nested paths are resolved correctly.
func TestResolveSourcePath(t *testing.T) {
	data := StepData{"a.b": map[string]interface{}{"c": 1}}
	val, ok := resolveSourcePath(data, "a.b.c")
	if !ok || val.(int) != 1 {
		t.Fatalf("unexpected result: %v %v", val, ok)
	}
}

// TestRAGPipeline runs the RAG pipeline end-to-end with a local HTTP server.
func TestRAGPipeline(t *testing.T) {
	srv := &http.Server{Addr: ":8080"}
	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in map[string]interface{}
		json.NewDecoder(r.Body).Decode(&in)
		json.NewEncoder(w).Encode(map[string]string{"completion": "test"})
	})
	go srv.ListenAndServe()
	defer srv.Shutdown(context.Background())

	store := vectorstore.NewMemoryStore()
	vectorstore.SetDefaultStore(store)
	emb := tools.BasicHashEmbed("hello", 128)
	store.Upsert(context.Background(), []vectorstore.Document{{ID: "1", Embedding: emb, Metadata: map[string]interface{}{"text": "hello"}}})

	pipeline := DefaultRAGPipeline("rag_test")
	orc := NewOrchestrator()
	input := map[string]interface{}{
		"query":               "hello",
		"template":            "{{range .documents}}{{.metadata.text}}{{end}}",
		"completion_endpoint": "http://localhost:8080",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, err := orc.ExecutePipeline(ctx, pipeline, input)
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	resp, ok := ExtractRAGResponse(data)
	if !ok || resp.Answer != "test" || len(resp.Documents) == 0 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}
