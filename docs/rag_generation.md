# RAG Generation Pipeline

This document outlines the current implementation for retrieval augmented generation (RAG).
The goal is to expose a minimal yet production ready chain that can be used for
early testing of end to end flows.

## Overview

1. **EmbeddingAgent** – Converts the user query into a vector using the configured
   `EmbeddingProvider`.
2. **RetrievalAgent** – Looks up similar documents from the `VectorStore`.
3. **RerankAgent** – Orders the retrieved documents by relevance score.
4. **ContextBuilderAgent** – Collates the reranked documents and optional extra
   context into a formatted string ready for prompting.
5. **PromptAgent** – Injects the retrieved documents, original query and
   formatted context into a templated prompt.
6. **GenerationAgent** – Sends the prompt to the Universal MCP endpoint and
   returns the completion text.
7. **Reasoning Step (optional)** – When enabled, the pipeline builds a second
   prompt containing the first answer and source context. Another
   `GenerationAgent` call produces a natural language explanation which is
   returned as part of the `RAGResponse`.

`internal/orchestrator.BuildRAGPipeline` wires these steps together. Callers may
provide `RAGPipelineOptions` to define defaults such as retrieval depth,
context formatting behaviour or the generation endpoint. The initial input must
include a user `query` and a prompt `template`. Optional fields include
`model`, `top_k`, `completion_endpoint`, `extra_context`,
`context_field`, `context_separator`, `context_max_chars` and
`reason_template`. When `EnableReasoning` is set in `RAGPipelineOptions`, the
`reason_template` is used to craft a second prompt for explanations. After
execution, `ExtractRAGResponse` converts the raw `StepData` into a
`RAGResponse` struct containing the original query, generated answer, the
formatted context string, reasoning text and the list of injected
`ContextDocument` values. A flag `context_truncated` indicates if the context
string was shortened to the requested length.

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
- **Reasoning step refinement** – develop prompt patterns and heuristics so the
  optional reasoning generation consistently explains how the answer was
  produced.
- **Response schemas** – formalise the structure returned by the pipeline so
  downstream services can consume answers and reasoning without ad-hoc parsing.
- **Full tracing** – expose optional step level traces in the `RAGResponse` for
  debugging and evaluation.
- **Context size heuristics** – dynamically tune `max_chars` based on model
  limits or user preferences.

These tasks will harden the pipeline for real workloads while keeping the
interfaces stable.
