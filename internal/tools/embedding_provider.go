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

var defaultEmbeddingProvider EmbeddingProvider

// SetDefaultEmbeddingProvider defines the global provider used when none is specified.
func SetDefaultEmbeddingProvider(p EmbeddingProvider) { defaultEmbeddingProvider = p }

// DefaultEmbeddingProvider returns the configured provider or a HashEmbeddingProvider.
func DefaultEmbeddingProvider() EmbeddingProvider {
	if defaultEmbeddingProvider == nil {
		defaultEmbeddingProvider = HashEmbeddingProvider{Dim: 128}
	}
	return defaultEmbeddingProvider
}
