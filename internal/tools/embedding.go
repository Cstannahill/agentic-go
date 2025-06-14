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

// EmbeddingTool produces embeddings for provided text. The actual embedding
// generation is delegated to an EmbeddingProvider so different backends can be
// swapped in as needed.
type EmbeddingTool struct {
	Provider EmbeddingProvider
}

// NewEmbeddingTool creates an EmbeddingTool using a HashEmbeddingProvider with
// the specified dimension. This maintains backwards compatibility with previous
// behaviour while allowing custom providers.
func NewEmbeddingTool(dim int) *EmbeddingTool {
	return NewEmbeddingToolWithProvider(HashEmbeddingProvider{Dim: dim})
}

// NewEmbeddingToolWithProvider allows callers to specify a custom provider such
// as a remote embedding service.
func NewEmbeddingToolWithProvider(p EmbeddingProvider) *EmbeddingTool {
	return &EmbeddingTool{Provider: p}
}

// Run implements the Tool interface.
func (e *EmbeddingTool) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	txt, _ := input["text"].(string)
	if txt == "" {
		return nil, errors.New("text field required")
	}

	if e.Provider == nil {
		e.Provider = DefaultEmbeddingProvider()
	}

	emb, err := e.Provider.Embed(ctx, txt)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"embedding": emb}, nil
}
