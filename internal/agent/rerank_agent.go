package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"agentic.example.com/mvp/internal/tools"
)

// RerankAgent orders documents based on score.
type RerankAgent struct {
	id   string
	tool *tools.RerankTool
}

func NewRerankAgent() *RerankAgent {
	return &RerankAgent{id: fmt.Sprintf("rerank-agent-%s", uuid.NewString()), tool: tools.NewRerankToolWithProvider(tools.DefaultRerankProvider())}
}

func (r *RerankAgent) ID() string { return r.id }

func (r *RerankAgent) Execute(ctx context.Context, task Task) Result {
	out, err := r.tool.Run(ctx, task.Input)
	if err != nil {
		return Result{TaskID: task.ID, Error: err}
	}
	return Result{TaskID: task.ID, Output: out, Successful: true}
}

func init() {
	Register("RerankAgent", func() Agent { return NewRerankAgent() })
}
