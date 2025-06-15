package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"agentic.example.com/mvp/internal/tools"
)

// IngestAgent embeds and stores documents in the configured vector store.
type IngestAgent struct {
	id   string
	tool *tools.IngestTool
}

// NewIngestAgent creates an agent with default tool configuration.
func NewIngestAgent() *IngestAgent {
	return &IngestAgent{
		id:   fmt.Sprintf("ingest-agent-%s", uuid.NewString()),
		tool: tools.NewIngestTool(),
	}
}

func (i *IngestAgent) ID() string { return i.id }

func (i *IngestAgent) Execute(ctx context.Context, task Task) Result {
	out, err := i.tool.Run(ctx, task.Input)
	if err != nil {
		return Result{TaskID: task.ID, Error: err}
	}
	return Result{TaskID: task.ID, Output: out, Successful: true}
}

func init() {
	Register("IngestAgent", func() Agent { return NewIngestAgent() })
}
