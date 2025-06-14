package config

import (
	"os"
)

// VectorStoreConfig holds configuration for connecting to a vector database.
type VectorStoreConfig struct {
	Endpoint   string
	Collection string
	APIKey     string
	Insecure   bool
}

// Config aggregates runtime settings for the pipeline tools.
type Config struct {
	VectorStore       VectorStoreConfig
	EmbeddingEndpoint string
	RerankEndpoint    string
}

// LoadFromEnv builds a Config from environment variables.
// VECTORSTORE_ENDPOINT and VECTORSTORE_COLLECTION must be set to enable remote store.
func LoadFromEnv() Config {
	insecure := false
	if os.Getenv("VECTORSTORE_INSECURE") == "1" {
		insecure = true
	}
	return Config{
		VectorStore: VectorStoreConfig{
			Endpoint:   os.Getenv("VECTORSTORE_ENDPOINT"),
			Collection: os.Getenv("VECTORSTORE_COLLECTION"),
			APIKey:     os.Getenv("VECTORSTORE_API_KEY"),
			Insecure:   insecure,
		},
		EmbeddingEndpoint: os.Getenv("EMBEDDING_ENDPOINT"),
		RerankEndpoint:    os.Getenv("RERANK_ENDPOINT"),
	}
}
