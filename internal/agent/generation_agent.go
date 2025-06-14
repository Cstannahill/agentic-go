package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"agentic.example.com/mvp/internal/tools"
)

// GenerationAgent calls a language model to generate a response from a prompt.
type GenerationAgent struct {
	id   string
	tool *tools.CompletionTool
}

// NewGenerationAgent creates a GenerationAgent with the given completion endpoint.
func NewGenerationAgent(endpoint string) *GenerationAgent {
	return &GenerationAgent{
		id:   fmt.Sprintf("generation-agent-%s", uuid.NewString()),
		tool: tools.NewCompletionTool(endpoint),
	}
}

// NewDefaultGenerationAgent constructs an agent using the default completion endpoint.
func NewDefaultGenerationAgent() *GenerationAgent {
	return NewGenerationAgent(tools.DefaultCompletionEndpoint())
}

func (g *GenerationAgent) ID() string { return g.id }

// Execute expects input with key "prompt" and optional "model".
// It forwards the request to the CompletionTool.
func (g *GenerationAgent) Execute(ctx context.Context, task Task) Result {
	tool := g.tool
	if ep, ok := task.Input["endpoint"].(string); ok && ep != "" && ep != g.tool.Endpoint {
		tool = tools.NewCompletionTool(ep)
	}
	out, err := tool.Run(ctx, task.Input)
	if err != nil {
		return Result{TaskID: task.ID, Error: err}
	}
	return Result{TaskID: task.ID, Output: out, Successful: true}
}

func init() {
	Register("GenerationAgent", func() Agent { return NewDefaultGenerationAgent() })
}
