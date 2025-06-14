# Vector Pipeline Components

This document describes the initial implementation of the embedding, storage,
retrieval and reranking pieces of the agentic pipeline.

## Overview

The pipeline exposes dedicated agents and tools to work with a vector store.
All data flows through the `vectorstore` package which now includes both an
in-memory implementation for tests and a `QdrantStore` that talks to a remote
Qdrant instance. The embedding step can be backed by a pluggable
`EmbeddingProvider` so that local hashing can easily be replaced with a real
model.

### Packages

* `internal/vectorstore` – Defines the `VectorStore` interface, `MemoryStore`
  for local use and `QdrantStore` for production deployments.
* `internal/tools` – Implements the common `Tool` interface:
  * `EmbeddingTool` – uses an `EmbeddingProvider` such as `HashEmbeddingProvider`
    or `RemoteEmbeddingProvider`.
  * `RetrievalTool` – queries a configured `VectorStore`.
  * `RerankTool` – orders documents by score.
* `internal/agent` – Agents wrapping the tools so they can run as pipeline steps:
  * `EmbeddingAgent`
  * `RetrievalAgent`
  * `RerankAgent`

## Remaining Work

1. **Authentication & TLS** – secure connections to the remote vector store and
   embedding service.
2. **Advanced Reranking** – integrate a cross-encoder model to score documents
   based on query relevance.
3. **Pipeline Configuration** – load store URLs and embedding options from
   external configuration files.
4. **Observability** – structured logging and metrics around all vector
   operations.

These steps will take the foundation here to a live-ready state while keeping the
API surface stable.
