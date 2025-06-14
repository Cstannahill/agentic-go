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
	if results[0].Score == 0 {
		t.Fatalf("expected score to be set")
	}

	if err := store.Delete(nil, []string{"1"}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	results, err = store.Query(nil, []float64{1, 0}, 1)
	if err != nil {
		t.Fatalf("query after delete: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected empty results after delete: %+v", results)
	}
}
