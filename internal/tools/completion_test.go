package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompletionTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if in["prompt"] != "hi" {
			t.Fatalf("unexpected prompt: %v", in)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"completion": "hello"})
	}))
	defer srv.Close()

	tool := NewCompletionTool(srv.URL)
	out, err := tool.Run(context.Background(), map[string]interface{}{"prompt": "hi"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if out["completion"] != "hello" {
		t.Fatalf("unexpected completion: %v", out)
	}
}
