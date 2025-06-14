# Vector Pipeline Components

This document describes the initial implementation of the embedding, storage,
retrieval and reranking pieces of the agentic pipeline.

## Overview

The pipeline now exposes dedicated agents and tools to work with a vector store.
All data is stored through the `vectorstore` package which currently provides
an in-memory implementation for local testing.  Each piece is designed to be
swappable with a production ready backend.

### Packages

* `internal/vectorstore` – Defines the `VectorStore` interface and a basic
  `MemoryStore` used for early testing.
* `internal/tools` – Contains tools implementing the common `Tool` interface:
  * `EmbeddingTool` – generates deterministic embeddings.
  * `RetrievalTool` – queries a `VectorStore` for similar documents.
  * `RerankTool` – orders documents by score.
* `internal/agent` – Agents wrapping the tools so they can run as pipeline steps:
  * `EmbeddingAgent`
  * `RetrievalAgent`
  * `RerankAgent`

## Remaining Work

1. **Persistent Vector Store** – replace `MemoryStore` with a production ready
   database (e.g. Qdrant or Weaviate) behind the same interface.
2. **Real Embedding Model** – integrate with an actual embedding service or
   model. The current hash-based embedding is deterministic but not meaningful.
3. **Scoring for Reranking** – the rerank agent assumes a `score` field. A
   real implementation should compute scores based on query context and document
   relevance.
4. **Configuration** – allow pipeline definitions to specify which vector store
   to use and embed model parameters.
5. **Metrics & Error Handling** – add logging and observability hooks around
   vector operations.

These steps will take the foundation here to a live-ready state while keeping the
API surface stable.
