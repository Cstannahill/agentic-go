# RAG Generation Pipeline

This document outlines the current implementation for retrieval augmented generation (RAG).
The goal is to expose a minimal yet production ready chain that can be used for
early testing of end to end flows.

## Overview

1. **EmbeddingAgent** – Converts the user query into a vector using the configured
   `EmbeddingProvider`.
2. **RetrievalAgent** – Looks up similar documents from the `VectorStore`.
3. **RerankAgent** – Orders the retrieved documents by relevance score.
4. **PromptAgent** – Injects the retrieved context into a templated prompt.
5. **GenerationAgent** – Sends the prompt to the Universal MCP endpoint and
   returns the completion text.

Each component runs as an agent so steps may execute concurrently where
possible.  The `PromptAgent` and `GenerationAgent` are new additions that move
the codebase beyond simple examples.

## Remaining Work

- **Real LLM integration** – the `GenerationAgent` currently posts to a
  configurable HTTP endpoint.  Wiring this up to the chosen model provider and
  handling authentication is required.
- **Streaming responses** – the completion API currently returns the full text at
  once.  Support for server-sent events or gRPC streaming will allow incremental
  delivery to clients.
- **Prompt templates from configuration** – templates are supplied in the task
  input today.  Loading and versioning them from external files is planned.
- **Observability and metrics** – structured logging of each step plus basic
  metrics (latency, failure counts) are needed before production use.

These tasks will harden the pipeline for real workloads while keeping the
interfaces stable.
