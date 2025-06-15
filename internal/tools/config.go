package tools

import "agentic.example.com/mvp/internal/config"

// InitDefaults configures global providers based on the given Config.
// When endpoints are empty the built-in providers remain in use.
func InitDefaults(cfg config.Config) {
	if cfg.EmbeddingEndpoint != "" {
		opts := []RemoteEmbedOption{}
		if cfg.EmbeddingAPIKey != "" {
			opts = append(opts, WithEmbedHeader("Authorization", cfg.EmbeddingAPIKey))
		}
		SetDefaultEmbeddingProvider(NewRemoteEmbeddingProvider(cfg.EmbeddingEndpoint, opts...))
	} else if cfg.EmbeddingDim > 0 {
		SetDefaultEmbeddingProvider(HashEmbeddingProvider{Dim: cfg.EmbeddingDim})
	}

	if cfg.RerankEndpoint != "" {
		opts := []RemoteRerankOption{}
		if cfg.RerankAPIKey != "" {
			opts = append(opts, WithRerankHeader("Authorization", cfg.RerankAPIKey))
		}
		SetDefaultRerankProvider(NewRemoteRerankProvider(cfg.RerankEndpoint, opts...))
	}
	if cfg.CompletionEndpoint != "" {
		SetDefaultCompletionEndpoint(cfg.CompletionEndpoint)
	}
	if cfg.RetrievalTopK > 0 {
		SetDefaultTopK(cfg.RetrievalTopK)
	}
}
