package vectorstore

import "testing"

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	SetDefaultStore(store)
	doc := Document{ID: "1", Embedding: []float64{1, 0}, Metadata: map[string]interface{}{"text": "a"}}
	if err := store.Upsert(nil, []Document{doc}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	results, err := store.Query(nil, []float64{1, 0}, 1)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(results) != 1 || results[0].ID != "1" {
		t.Fatalf("unexpected results: %+v", results)
	}
}
