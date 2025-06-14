package tools

import "agentic.example.com/mvp/internal/config"

// InitDefaults configures global providers based on the given Config.
// When endpoints are empty the built-in providers remain in use.
func InitDefaults(cfg config.Config) {
	if cfg.EmbeddingEndpoint != "" {
		SetDefaultEmbeddingProvider(NewRemoteEmbeddingProvider(cfg.EmbeddingEndpoint))
	} else if cfg.EmbeddingDim > 0 {
		SetDefaultEmbeddingProvider(HashEmbeddingProvider{Dim: cfg.EmbeddingDim})
	}

	if cfg.RerankEndpoint != "" {
		SetDefaultRerankProvider(NewRemoteRerankProvider(cfg.RerankEndpoint))
	}

	if cfg.RetrievalTopK > 0 {
		SetDefaultTopK(cfg.RetrievalTopK)
	}
}
