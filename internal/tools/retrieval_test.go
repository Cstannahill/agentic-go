package tools

import (
	"testing"

	"agentic.example.com/mvp/internal/vectorstore"
)

func TestRetrievalTool(t *testing.T) {
	store := vectorstore.NewMemoryStore()
	vectorstore.SetDefaultStore(store)
	emb := BasicHashEmbed("hello", 8)
	store.Upsert(nil, []vectorstore.Document{{ID: "1", Embedding: emb, Metadata: map[string]interface{}{"tag": "x"}}})
	tool := NewRetrievalTool(store, 1)
	out, err := tool.Run(nil, map[string]interface{}{"embedding": emb})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	docs := out["documents"].([]map[string]interface{})
	if len(docs) != 1 || docs[0]["id"] != "1" {
		t.Fatalf("unexpected result: %+v", out)
	}
	if _, ok := docs[0]["score"]; !ok {
		t.Fatalf("missing score: %+v", docs[0])
	}

	out, err = tool.Run(nil, map[string]interface{}{"embedding": emb, "filter": map[string]interface{}{"tag": "x"}})
	if err != nil || len(out["documents"].([]map[string]interface{})) != 1 {
		t.Fatalf("filter not applied: %v %v", err, out)
	}

	out, err = tool.Run(nil, map[string]interface{}{"embedding": emb, "filter": map[string]interface{}{"tag": "y"}})
	if err != nil {
		t.Fatalf("run with filter: %v", err)
	}
	if len(out["documents"].([]map[string]interface{})) != 0 {
		t.Fatalf("expected empty result with non matching filter")
	}
}
