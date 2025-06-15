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
remote cross-encoder service. Stores and providers can now be initialised from
environment configuration using `setup.InitFromEnv`. This convenience
function loads configuration and wires up `vectorstore.InitDefault` and
`tools.InitDefaults` in one call.
Remote providers include simple retry logic so transient network failures are
automatically retried.

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

The `QdrantStore` constructor accepts options for API keys, TLS behaviour and a
custom `http.Client` so deployments can tune connection settings.

### Recent Additions

- `VectorStore.Query` now accepts a `QueryRequest` struct with optional
  metadata `Filter` enabling server side filtering.
- `RemoteEmbeddingProvider` and `RemoteRerankProvider` support custom HTTP
  headers (e.g. API tokens) and use exponential backoff on retry.
- Configuration variables `EMBEDDING_API_KEY` and `RERANK_API_KEY` pass these
  tokens to the providers.
- `setup.InitFromEnv` simplifies bootstrapping by wiring providers and the
  vector store from environment variables.
- The RAG pipeline's retrieval step now accepts an optional `filter` map which
  is forwarded to the vector database query.

## Remaining Work

The following tasks will harden the pipeline so it can service live traffic
while keeping the public API stable.

1. **Authentication & TLS** – secure connections to the remote vector store and
   embedding/ rerank services. Basic API token support is in place but
   certificate validation and OAuth flows still need wiring up.
2. **Advanced Reranking** – integrate a cross-encoder model to score documents
   based on query relevance. The `RemoteRerankProvider` is a placeholder for
   this.
3. **Observability** – add structured logging and Prometheus metrics around all
   vector operations.
4. **Dataset Management** – the `IngestTool` provides a simple path for adding
   documents, but bulk import and update workflows are still needed.
5. **Configuration Loader** – expose helper functions to read YAML/JSON configs
   so environments can be provisioned without recompilation.
6. **Production Configuration** – tune embedding dimension and retrieval depth
   via `EMBEDDING_DIM` and `RETRIEVAL_TOP_K` environment variables. This allows
   consistent behaviour across deployments.
7. **Integration Tests** – expand test coverage by running the pipeline against
   ephemeral Qdrant and rerank containers.
8. **Operational Guides** – document deployment topologies and example
   configuration files to assist early testers.
9. **Failure Handling** – surface structured errors from tools and implement
   retry logic around transient network failures beyond the current best-effort
   approach.
