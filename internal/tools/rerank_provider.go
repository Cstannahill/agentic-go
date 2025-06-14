package tools

import "context"

// RerankProvider assigns relevance scores to documents given a query.
type RerankProvider interface {
	Rerank(ctx context.Context, query string, docs []string) ([]float64, error)
}

// DefaultRerankProvider is used when no provider is specified.
var defaultRerankProvider RerankProvider

// SetDefaultRerankProvider sets the global provider used by RerankTool.
func SetDefaultRerankProvider(p RerankProvider) { defaultRerankProvider = p }

// DefaultRerankProvider returns the configured provider or nil.
func DefaultRerankProvider() RerankProvider { return defaultRerankProvider }
