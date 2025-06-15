# Runtime Configuration

The vector pipeline relies on a few environment variables to configure remote services. When none are provided, an in-memory vector store and local hash embedding provider are used.

| Variable | Purpose |
| --- | --- |
| `VECTORSTORE_ENDPOINT` | Base URL for a Qdrant instance. |
| `VECTORSTORE_COLLECTION` | Name of the collection to use. |
| `VECTORSTORE_API_KEY` | Optional API key for Qdrant. |
| `VECTORSTORE_INSECURE` | Set to `1` to skip TLS verification. |
| `EMBEDDING_ENDPOINT` | HTTP endpoint for generating embeddings. |
| `RERANK_ENDPOINT` | HTTP endpoint for reranking documents. |
| `EMBEDDING_API_KEY` | Optional auth token for the embedding service. |
| `RERANK_API_KEY` | Optional auth token for the rerank service. |
| `COMPLETION_ENDPOINT` | HTTP endpoint for language model completion. |
| `EMBEDDING_DIM` | Dimension for the built-in hash embedding provider. |
| `RETRIEVAL_TOP_K` | Default number of documents to retrieve. |
Applications can load these settings via `config.LoadFromEnv()` and use them to initialise the default tools and stores.
