package config

import (
	"os"
	"strconv"
)

// VectorStoreConfig holds configuration for connecting to a vector database.
type VectorStoreConfig struct {
	Endpoint   string
	Collection string
	APIKey     string
	Insecure   bool
}

// Config aggregates runtime settings for the pipeline tools.
// Config aggregates runtime settings for the pipeline tools.
// Values may be empty when the corresponding environment variables are unset.
type Config struct {
	VectorStore        VectorStoreConfig
	EmbeddingEndpoint  string
	RerankEndpoint     string
	CompletionEndpoint string
	EmbeddingDim       int
	RetrievalTopK      int
}

// LoadFromEnv builds a Config from environment variables.
// VECTORSTORE_ENDPOINT and VECTORSTORE_COLLECTION must be set to enable remote store.
func LoadFromEnv() Config {
	insecure := false
	if os.Getenv("VECTORSTORE_INSECURE") == "1" {
		insecure = true
	}

	embDim := 0
	if v := os.Getenv("EMBEDDING_DIM"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			embDim = n
		}
	}

	topK := 0
	if v := os.Getenv("RETRIEVAL_TOP_K"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			topK = n
		}
	}

	return Config{
		VectorStore: VectorStoreConfig{
			Endpoint:   os.Getenv("VECTORSTORE_ENDPOINT"),
			Collection: os.Getenv("VECTORSTORE_COLLECTION"),
			APIKey:     os.Getenv("VECTORSTORE_API_KEY"),
			Insecure:   insecure,
		},
		EmbeddingEndpoint:  os.Getenv("EMBEDDING_ENDPOINT"),
		RerankEndpoint:     os.Getenv("RERANK_ENDPOINT"),
		CompletionEndpoint: os.Getenv("COMPLETION_ENDPOINT"),
		EmbeddingDim:       embDim,
		RetrievalTopK:      topK,
	}
}
