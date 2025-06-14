# Universal MCP Layer

The **Master Control Program (MCP)** provides a central gateway for executing remote models and tools regardless of where they live. It standardizes how the orchestrator communicates with external services.

## Purpose

- **Unify Remote Calls** – Whether invoking a local Python script, a containerized service or a cloud LLM, the MCP exposes a consistent HTTP interface.
- **Language Agnostic** – Clients written in any language send a task description and receive structured results. Internally the MCP proxies the request to the correct backend.
- **Concurrent Handling** – The MCP is written in Go so each inbound request is handled in its own goroutine. Long‐running model invocations cannot block others.

## Core Features

1. **Adapter Plugins**
    - Each external tool is wrapped by an adapter implementing a common interface.
    - Adapters translate generic task payloads into the target system’s API format and back again.
2. **Routing Logic**
    - Requests specify the desired adapter by name. The MCP routes the payload to the matching adapter.
    - Defaults and fallbacks allow one tool to substitute for another when needed.
3. **Streaming Support**
    - For LLMs or tools that produce incremental output, adapters may stream results over HTTP using server‑sent events or WebSockets.
4. **Observability Hooks**
    - Every request/response pair is logged with timing and optional trace IDs.
    - Metrics can be exported to Prometheus or other monitoring systems.
5. **Auth and Rate Limiting**
    - Basic authentication and per‑adapter rate limits protect remote services from abuse.

## Usage Within the Orchestrator

The orchestrator’s `HTTPCallAgent` interacts with the MCP rather than calling tools directly. Pipelines simply specify which MCP adapter to use and supply any necessary parameters. This keeps step definitions clean and allows tools to be swapped without changing pipeline code.

```yaml
# Example pipeline snippet referencing an MCP adapter
- step_name: generate_answer
  agent_type: HTTPCallAgent
  config:
    url: http://localhost:8000/mcp/invoke/claude
    method: POST
    payload_template:
      query: "$.initial_input.user_query"
```

## Roadmap

- **Adapter Library** – Build adapters for common model providers (Claude, OpenAI, local embeddings engines).
- **Central Configuration** – YAML or JSON file mapping adapter names to their endpoints and auth credentials.
- **Caching** – Optional result caching to speed up repeated requests.
- **CLI Utilities** – Commands for registering new adapters and health checking.

The MCP layer serves as the universal glue between the Go orchestration engine and any external capability. Adding a new model becomes a matter of writing a small adapter while the core pipeline logic remains unchanged.

