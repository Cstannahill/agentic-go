# Vector Pipeline Components

This document describes the initial implementation of the embedding, storage,
retrieval and reranking pieces of the agentic pipeline.

## Overview

The pipeline exposes dedicated agents and tools to work with a vector store.
All data flows through the `vectorstore` package which now includes both an
in-memory implementation for tests and a `QdrantStore` that talks to a remote
Qdrant instance. The embedding step can be backed by a pluggable
`EmbeddingProvider` so that local hashing can easily be replaced with a real
model. Reranking likewise uses a pluggable `RerankProvider` which may call a
remote cross-encoder service.  Stores and providers can now be initialised from
environment configuration using `config.LoadFromEnv` together with
`vectorstore.InitDefault` and `tools.InitDefaults`.

The default hash embedding dimension and retrieval depth are also
configurable through `EMBEDDING_DIM` and `RETRIEVAL_TOP_K`. These values
allow tuning relevance without code changes.

### Packages

* `internal/vectorstore` – Defines the `VectorStore` interface, `MemoryStore`
  for local use and `QdrantStore` for production deployments.
* `internal/tools` – Implements the common `Tool` interface:
  * `EmbeddingTool` – uses a pluggable `EmbeddingProvider` (`HashEmbeddingProvider`,
    `RemoteEmbeddingProvider`, etc.).
  * `RetrievalTool` – queries a configured `VectorStore`.
  * `RerankTool` – orders documents using a `RerankProvider` when available.
  * `IngestTool` – embeds text and stores it in the configured `VectorStore`.
* `internal/agent` – Agents wrapping the tools so they can run as pipeline steps:
  * `EmbeddingAgent`
  * `RetrievalAgent`
  * `RerankAgent`
  * `IngestAgent`

## Remaining Work

The following items track what is still required before the vector pipeline can
be considered production ready:

1. **Authentication & TLS** – secure connections to the remote vector store and
   remote providers. Qdrant API key support has landed but certificate
   validation and token based auth need wiring up.
2. **Advanced Reranking** – integrate a cross-encoder model to score documents
   based on query relevance. The `RemoteRerankProvider` is a placeholder for this.
3. **Observability** – add structured logging and Prometheus metrics around all
   vector operations.
4. **Dataset Management** – the new `IngestTool` and `IngestAgent` provide a
   simple path for adding documents, but bulk import and update workflows are
   still needed.
5. **Configuration Loader** – expose helper functions to read YAML/JSON configs
   so environments can be provisioned without recompilation.
6. **Production Configuration** – tune embedding dimension and retrieval depth
   via `EMBEDDING_DIM` and `RETRIEVAL_TOP_K` environment variables. This allows
   consistent behaviour across deployments.
7. **Integration Tests** – add test suites exercising the Qdrant client and
   remote rerank service using local containers.
8. **Error Handling & Retry** – provide clearer error types and automatic
   retries for transient failures.

Addressing these areas will harden the pipeline while keeping the API surface
stable for early testing.
