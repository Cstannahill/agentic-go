package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRerankTool(t *testing.T) {
	tool := NewRerankTool()
	docs := []map[string]interface{}{{"id": "a", "score": 0.1}, {"id": "b", "score": 0.9}}
	out, err := tool.Run(nil, map[string]interface{}{"documents": docs})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	r := out["reranked"].([]map[string]interface{})
	if r[0]["id"] != "b" {
		t.Fatalf("unexpected order: %+v", r)
	}
}

func TestRemoteRerankProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"scores": []float64{0.2, 0.8}})
	}))
	defer srv.Close()

	p := NewRemoteRerankProvider(srv.URL)
	scores, err := p.Rerank(context.Background(), "hi", []string{"a", "b"})
	if err != nil {
		t.Fatalf("rerank: %v", err)
	}
	if len(scores) != 2 || scores[1] <= scores[0] {
		t.Fatalf("unexpected scores: %v", scores)
	}
}

func TestRemoteRerankProviderRetry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			http.Error(w, "bad", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"scores": []float64{0.5}})
	}))
	defer srv.Close()

	p := NewRemoteRerankProvider(srv.URL)
	p.MaxRetries = 1
	scores, err := p.Rerank(context.Background(), "q", []string{"a"})
	if err != nil {
		t.Fatalf("retry rerank: %v", err)
	}
	if len(scores) != 1 || scores[0] != 0.5 || attempts != 2 {
		t.Fatalf("unexpected result %v attempts=%d", scores, attempts)
	}
}
