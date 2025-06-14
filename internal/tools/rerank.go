package tools

import (
	"context"
	"errors"
	"sort"
)

// RerankTool sorts documents based on a provided score field.
// RerankTool orders documents based on scores from a provider or existing score field.
type RerankTool struct {
	Provider RerankProvider
}

// NewRerankTool creates a simple RerankTool.
func NewRerankTool() *RerankTool { return &RerankTool{} }

// NewRerankToolWithProvider allows callers to specify a custom provider.
func NewRerankToolWithProvider(p RerankProvider) *RerankTool { return &RerankTool{Provider: p} }

// Run implements the Tool interface. Input expects `documents` as slice of maps
// with a numeric `score` field.
func (r *RerankTool) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	docs, ok := input["documents"].([]map[string]interface{})
	if !ok {
		return nil, errors.New("documents must be provided")
	}
	query, _ := input["query"].(string)
	if r.Provider == nil {
		r.Provider = DefaultRerankProvider()
	}
	if r.Provider != nil && query != "" {
		texts := make([]string, len(docs))
		for i, d := range docs {
			if t, ok := d["text"].(string); ok {
				texts[i] = t
			}
		}
		if scores, err := r.Provider.Rerank(ctx, query, texts); err == nil && len(scores) == len(docs) {
			for i := range docs {
				docs[i]["score"] = scores[i]
			}
		}
	}
	sort.Slice(docs, func(i, j int) bool {
		si, _ := docs[i]["score"].(float64)
		sj, _ := docs[j]["score"].(float64)
		return si > sj
	})
	return map[string]interface{}{"reranked": docs}, nil
}
