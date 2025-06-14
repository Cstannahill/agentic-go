package agent

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/google/uuid"
)

// PromptAgent builds a prompt by applying a Go template to provided context.
type PromptAgent struct {
	id string
}

// NewPromptAgent creates a PromptAgent.
func NewPromptAgent() *PromptAgent {
	return &PromptAgent{id: fmt.Sprintf("prompt-agent-%s", uuid.NewString())}
}

func (p *PromptAgent) ID() string { return p.id }

// Execute expects input with keys:
//
//	template: string - Go text/template string
//	context:  map[string]interface{} - data for template execution
//
// Returns a map with key "prompt" containing the rendered template.
func (p *PromptAgent) Execute(ctx context.Context, task Task) Result {
	tmplStr, _ := task.Input["template"].(string)
	if tmplStr == "" {
		return Result{TaskID: task.ID, Error: fmt.Errorf("template required")}
	}
	data, _ := task.Input["context"].(map[string]interface{})
	tmpl, err := template.New("prompt").Parse(tmplStr)
	if err != nil {
		return Result{TaskID: task.ID, Error: err}
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return Result{TaskID: task.ID, Error: err}
	}
	return Result{TaskID: task.ID, Output: map[string]interface{}{"prompt": buf.String()}, Successful: true}
}

func init() {
	Register("PromptAgent", func() Agent { return NewPromptAgent() })
}
