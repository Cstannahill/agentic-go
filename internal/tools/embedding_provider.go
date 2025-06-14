package tools

import "context"

// EmbeddingProvider defines how text is converted to a vector representation.
type EmbeddingProvider interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}

// HashEmbeddingProvider implements EmbeddingProvider using the BasicHashEmbed
// function. It is mainly intended for testing and local development.
type HashEmbeddingProvider struct{ Dim int }

func (h HashEmbeddingProvider) Embed(ctx context.Context, text string) ([]float64, error) {
	return BasicHashEmbed(text, h.Dim), nil
}
