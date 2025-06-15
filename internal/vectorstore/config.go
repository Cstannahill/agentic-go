package vectorstore

import "agentic.example.com/mvp/internal/config"

// NewFromConfig creates a VectorStore based on the provided configuration.
// If Endpoint is empty a MemoryStore is returned.
func NewFromConfig(cfg config.VectorStoreConfig) VectorStore {
	if cfg.Endpoint == "" {
		return NewMemoryStore()
	}
	opts := []QdrantOption{}
	if cfg.APIKey != "" {
		opts = append(opts, WithAPIKey(cfg.APIKey))
	}
	if cfg.Insecure {
		opts = append(opts, WithInsecureSkipVerify())
	}
	return NewQdrantStore(cfg.Endpoint, cfg.Collection, opts...)
}

// InitDefault sets the default global store using NewFromConfig.
func InitDefault(cfg config.VectorStoreConfig) {
	SetDefaultStore(NewFromConfig(cfg))
}
