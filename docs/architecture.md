# Agentic Orchestration Architecture

This document captures the overall design of the extensible Go orchestration layer.  The ideas here originate from `concept.md` and have been refined into a more concise form.

## Goals

- **Plug and play** – tools written in any language can be wrapped as an agent and used without knowing the rest of the system.
- **Concurrency everywhere** – goroutines and channels drive all coordination so the engine fully utilizes Go's strengths.
- **Composable pipelines** – each pipeline is a sequence of steps grouped into stages that can run concurrently.

## Core Components

### Agents

An agent is any piece of logic implementing the `Execute` method defined in [`agent.go`](../internal/agent/agent.go).  The orchestrator creates agents dynamically based on the `AgentType` specified in each step.  Built‑in examples include:

- `EchoAgent` – echos input and demonstrates timing control.
- `HTTPCallAgent` – performs HTTP requests and forwards the response.

Agents register themselves in a central registry so pipelines are decoupled from implementations.

### Pipeline Structure

Pipelines consist of ordered **groups** of steps.  Steps within a group run concurrently, each in its own goroutine, and send their results back through a channel.  The orchestrator waits for all steps in a group to finish before moving on to the next group.

```go
pipeline := orchestrator.Pipeline{
    ID: "example",
    Groups: []orchestrator.PipelineGroup{
        {Name: "fetch", Steps: []orchestrator.PipelineStep{/* ... */}},
        {Name: "process", Steps: []orchestrator.PipelineStep{/* ... */}},
    },
}
```

### Concurrency Model

Each step result is sent over a channel to the orchestrator.  This isolation makes it straightforward to introduce fan‑in or fan‑out patterns as the engine grows.  The design emphasises using goroutines and channels *any time possible* to keep the system responsive and scalable.

### Extensibility

Agents can live in separate repositories or services.  As long as they expose an HTTP interface or a small Go wrapper, they are usable by the engine.  This makes the orchestration layer language agnostic while still benefiting from Go's runtime.

### Universal MCP Layer

The **Master Control Program (MCP)** acts as a consolidated gateway for all remote models and tools. Instead of calling services directly, agents contact the MCP which routes each request to the correct adapter. See [universal_mcp.md](universal_mcp.md) for a complete design. The MCP provides:

- A uniform HTTP interface for invoking external capabilities.
- Adapter plugins that translate between generic task payloads and model-specific APIs.
- Concurrent handling so heavy model calls do not block others.
- Logging and metrics around every request.

## Planned Features and Roadmap

1. **HTTP Service** – expose pipeline execution through an API so other languages can dispatch tasks.
2. **Task State Tracking** – persist intermediate results for debugging and retrieval via API.
3. **Additional Agent Types**
   - Retrieval and embedding agents for working with vector stores.
   - Reranking and document attachment agents.
4. **Observability Improvements** – structured logging, tracing and metrics around each step.
5. **Configuration Loading** – ability to define pipelines and agent parameters from YAML or JSON files.

These items build on the skeleton currently in the repository and align with the direction outlined in `concept.md`.
