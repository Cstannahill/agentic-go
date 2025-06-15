# RAG Generation Pipeline

This document outlines the current implementation for retrieval augmented generation (RAG).
The goal is to expose a minimal yet production ready chain that can be used for
early testing of end to end flows.

## Overview

1. **EmbeddingAgent** – Converts the user query into a vector using the configured
   `EmbeddingProvider`.
2. **RetrievalAgent** – Looks up similar documents from the `VectorStore`.
3. **RerankAgent** – Orders the retrieved documents by relevance score.
4. **PromptAgent** – Injects the retrieved documents, original query and any
   extra context into a templated prompt.
5. **GenerationAgent** – Sends the prompt to the Universal MCP endpoint and
   returns the completion text.

`internal/orchestrator.BuildRAGPipeline` wires these steps together. Callers may
provide `RAGPipelineOptions` to define defaults such as retrieval depth or the
generation endpoint. The initial input must include a user `query` and a prompt
`template`. Optional fields include `model`, `top_k`, `completion_endpoint` and
`extra_context`. After execution, `ExtractRAGResponse` converts the raw
`StepData` into a `RAGResponse` struct containing the original query, generated
answer and the list of injected `ContextDocument` values.

Each component runs as an agent so steps may execute concurrently where
possible.  The `PromptAgent` and `GenerationAgent` now accept runtime options
for the retrieval depth and completion endpoint allowing early integration tests
against real services.

## Remaining Work

- **Real LLM integration** – the `GenerationAgent` can point to any HTTP
  endpoint but proper authentication, retry logic and error handling still need
  to be implemented.
- **Failure handling and retries** – implement retry logic and propagate
  structured errors up the pipeline.
- **Streaming responses** – the completion API currently returns the full text at
  once.  Support for server-sent events or gRPC streaming will allow incremental
  delivery to clients.
- **Prompt templates from configuration** – templates are supplied in the task
  input today. Loading and versioning them from external files is planned.
- **Central configuration** – environment variables or files should define
  default endpoints and retrieval parameters so deployments remain consistent.
- **Observability and metrics** – structured logging of each step plus basic
  metrics (latency, failure counts) are needed before production use.
- **Advanced prompt management** – reference templates by name and version to
  allow consistent updates across deployments.
- **Better context filtering** – improved similarity scoring and heuristics will
  ensure only the most relevant documents are injected into the prompt.

These tasks will harden the pipeline for real workloads while keeping the
interfaces stable.
