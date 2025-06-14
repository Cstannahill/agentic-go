agentic-go/
├── cmd/
│   ├── agentrunner/        # Example application executing a pipeline
│   │   └── main.go
│   └── server/             # HTTP server exposing pipeline execution
│       └── main.go
├── internal/
│   ├── agent/              # Core agent interface and implementations
│   │   ├── agent.go
│   │   ├── http_agent.go
│   │   └── registry.go
│   ├── orchestrator/       # Pipeline and orchestrator logic
│   │   ├── orchestrator.go
│   │   └── pipeline.go
│   └── tools/              # Stubs for embedding, retrieval, reranking
│       ├── embedding.go
│       ├── rerank.go
│       ├── retrieval.go
│       └── tool.go
├── docs/                   # Project documentation
│   ├── architecture.md
│   └── universal_mcp.md
├── concept.md              # Original notes and design discussion
├── README.md
├── structure.md            # This file
├── go.mod
└── go.sum
