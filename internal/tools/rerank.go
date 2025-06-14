package tools

import (
	"context"
	"errors"
	"sort"
)

// RerankTool sorts documents based on a provided score field.
type RerankTool struct{}

// NewRerankTool creates a simple RerankTool.
func NewRerankTool() *RerankTool { return &RerankTool{} }

// Run implements the Tool interface. Input expects `documents` as slice of maps
// with a numeric `score` field.
func (r *RerankTool) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	docs, ok := input["documents"].([]map[string]interface{})
	if !ok {
		return nil, errors.New("documents must be provided")
	}
	sort.Slice(docs, func(i, j int) bool {
		si, _ := docs[i]["score"].(float64)
		sj, _ := docs[j]["score"].(float64)
		return si > sj
	})
	return map[string]interface{}{"reranked": docs}, nil
}
