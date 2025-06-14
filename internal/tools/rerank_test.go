package tools

import "testing"

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
