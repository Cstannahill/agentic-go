# Agentic Go Orchestration Skeleton

This repository experiments with an extensible orchestration layer for building agent based pipelines in Go.  The design focuses on using goroutines and channels wherever possible so that every step of a pipeline can run concurrently.

See [docs/architecture.md](docs/architecture.md) for a deeper architectural overview and the project roadmap. The
[Universal MCP Layer](docs/universal_mcp.md) document explains how remote models are unified behind a single gateway.
Runtime options for connecting to external services are described in [docs/configuration.md](docs/configuration.md).

## Features

- **Agent Registry** – Agents register by name and can be created on demand.  Pipelines are therefore decoupled from concrete implementations.
- **Pipeline Orchestrator** – Executes ordered groups of steps.  Steps within a group run in their own goroutine and communicate results via channels.
- **Tool Interfaces** – Stubs for embedding, retrieval and reranking demonstrate how external capabilities plug in.
- **HTTP Server Example** – `cmd/server` exposes pipeline execution through a simple API.
- **Configurable Retrieval** – `EMBEDDING_DIM` and `RETRIEVAL_TOP_K` environment
  variables allow tuning default embedding size and number of documents
  returned without recompiling.
- **Data Transform Agent** – Performs basic string manipulation operations.
- **Ingest Agents and Tools** – Easily embed and store new documents in the
  configured vector database.
- **RAG Generation Pipeline** – Embedding, retrieval, context formatting and generation agents ready for early testing.
- **Pipeline Builder** – `BuildRAGPipeline` constructs a ready-to-run pipeline with optional defaults, and
  `ExtractRAGResponse` transforms raw results into a structured `RAGResponse`.

## Example

See `cmd/agentrunner/main.go` for a basic pipeline using the `EchoAgent`. The orchestrator exposes both synchronous and asynchronous execution styles. `RunPipeline` streams step results over a channel so callers can react as work completes.

Additional agents can be created and registered to extend the system:

`cmd/agentrunner/main.go` defines a small pipeline using the `EchoAgent` and an `HTTPCallAgent`:

```go
func init() {
    agent.Register("MyAgent", func() agent.Agent { return NewMyAgent() })
}
```

Pipelines reference agents by name and the orchestrator will instantiate them at runtime using the registry.  Each step's output becomes available for later steps through a shared `StepData` map.

### Example Pipeline with `DataTransformAgent`

The `DataTransformAgent` manipulates text according to an `operation` value.
Supported operations are `uppercase`, `lowercase`, `reverse`, and `title`.

```go
pipeline := orchestrator.Pipeline{
    ID: "string_pipeline",
    Groups: []orchestrator.PipelineGroup{
        {
            Name: "transform",
            Steps: []orchestrator.PipelineStep{
                {
                    Name:      "make_upper",
                    AgentType: "DataTransformAgent",
                    AgentConfig: agent.Task{Description: "Uppercase the text"},
                    InputMappings: map[string]string{
                        "text":      "initial.input_text",
                        "operation": "initial.op",
                    },
                },
            },
        },
    },
}
```
