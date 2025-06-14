package tools

import (
	"context"
	"errors"
	"hash/fnv"
	"strings"
)

// BasicHashEmbed generates a deterministic embedding vector of fixed dimension.
func BasicHashEmbed(text string, dim int) []float64 {
	vec := make([]float64, dim)
	for _, w := range strings.Fields(strings.ToLower(text)) {
		h := fnv.New32a()
		h.Write([]byte(w))
		idx := int(h.Sum32()) % dim
		vec[idx] += 1
	}
	return vec
}

// EmbeddingTool produces embeddings for provided text.
type EmbeddingTool struct {
	Dim int
}

// NewEmbeddingTool returns an EmbeddingTool with the given dimension.
func NewEmbeddingTool(dim int) *EmbeddingTool { return &EmbeddingTool{Dim: dim} }

// Run implements the Tool interface.
func (e *EmbeddingTool) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	txt, _ := input["text"].(string)
	if txt == "" {
		return nil, errors.New("text field required")
	}
	emb := BasicHashEmbed(txt, e.Dim)
	return map[string]interface{}{"embedding": emb}, nil
}
