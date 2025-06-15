package setup

import (
	"agentic.example.com/mvp/internal/config"
	"agentic.example.com/mvp/internal/tools"
	"agentic.example.com/mvp/internal/vectorstore"
)

// InitFromEnv loads configuration from environment variables and
// initialises the default vector store and tool providers.
// It returns the resulting Config for further use by the caller.
func InitFromEnv() config.Config {
	cfg := config.LoadFromEnv()
	vectorstore.InitDefault(cfg.VectorStore)
	tools.InitDefaults(cfg)
	return cfg
}
