# Agentic Go Orchestration Skeleton

This repository contains an experimental orchestration layer for building agent based pipelines in Go. The goal is to provide a plug–and–play framework that can host any type of agent while leveraging Go's concurrency primitives.

## Features

- **Agent Registry** – Agents can register themselves by name so pipelines remain decoupled from concrete implementations.
- **Pipeline Orchestrator** – Executes a series of steps where each step invokes an agent with mapped inputs.
- **Concurrency by Default** – Agents run in their own goroutine and communicate results via channels.
- **Tool Interfaces** – Basic stubs for tools like embedding, retrieval and reranking are provided for future expansion.

## Example

See `cmd/agentrunner/main.go` for a basic pipeline using the `EchoAgent`. Additional agents can be created and registered to extend the system:

```go
func init() {
    agent.Register("MyAgent", func() agent.Agent { return NewMyAgent() })
}
```

Pipelines reference agents by name and the orchestrator will instantiate them at runtime using the registry.

