package tools

import "testing"
import "net/http"
import "net/http/httptest"
import "encoding/json"
import "context"

func TestBasicHashEmbed(t *testing.T) {
	v1 := BasicHashEmbed("hello world", 10)
	v2 := BasicHashEmbed("hello world", 10)
	if len(v1) != 10 || len(v2) != 10 {
		t.Fatalf("unexpected dimension")
	}
	for i := range v1 {
		if v1[i] != v2[i] {
			t.Fatalf("embedding not deterministic")
		}
	}
}

func TestEmbeddingToolWithProvider(t *testing.T) {
	tool := NewEmbeddingToolWithProvider(HashEmbeddingProvider{Dim: 6})
	out, err := tool.Run(nil, map[string]interface{}{"text": "foo bar"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	emb := out["embedding"].([]float64)
	if len(emb) != 6 {
		t.Fatalf("unexpected dimension %d", len(emb))
	}
}

func TestRemoteEmbeddingProvider(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"embedding": []float64{1, 2, 3}})
	}))
	defer srv.Close()

	p := NewRemoteEmbeddingProvider(srv.URL)
	emb, err := p.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatalf("embed: %v", err)
	}
	if len(emb) != 3 || emb[0] != 1 {
		t.Fatalf("unexpected embedding: %v", emb)
	}
}

func TestRemoteEmbeddingProviderRetry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			http.Error(w, "bad", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"embedding": []float64{9}})
	}))
	defer srv.Close()

	p := NewRemoteEmbeddingProvider(srv.URL)
	p.MaxRetries = 1
	emb, err := p.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatalf("embed retry: %v", err)
	}
	if len(emb) != 1 || emb[0] != 9 || attempts != 2 {
		t.Fatalf("unexpected result %+v attempts=%d", emb, attempts)
	}
}
