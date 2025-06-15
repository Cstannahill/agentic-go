package agent

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/google/uuid"
)

// PromptAgent builds a prompt by applying a Go template to provided data.
// It can merge arbitrary context with retrieved documents and the original query.
type PromptAgent struct {
	id string
}

// NewPromptAgent creates a PromptAgent with a unique ID.
func NewPromptAgent() *PromptAgent {
	return &PromptAgent{id: fmt.Sprintf("prompt-agent-%s", uuid.NewString())}
}

func (p *PromptAgent) ID() string { return p.id }

// Execute expects input keys:
//
//	template  - Go text/template string used to render the prompt
//	documents - []map[string]interface{} representing retrieved context
//	query     - optional original user query
//	context   - optional additional map merged into the template data
//
// The resulting prompt string is returned in the "prompt" field of the output map.
func (p *PromptAgent) Execute(ctx context.Context, task Task) Result {
	tmplStr, _ := task.Input["template"].(string)
	if tmplStr == "" {
		return Result{TaskID: task.ID, Error: fmt.Errorf("template required")}
	}

	data := map[string]interface{}{}
	if extra, ok := task.Input["context"].(map[string]interface{}); ok {
		for k, v := range extra {
			data[k] = v
		}
	}
	if docs, ok := task.Input["documents"].([]map[string]interface{}); ok {
		data["documents"] = docs
	}
	if q, ok := task.Input["query"].(string); ok {
		data["query"] = q
	}
	if a, ok := task.Input["answer"].(string); ok {
		data["answer"] = a
	}

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
