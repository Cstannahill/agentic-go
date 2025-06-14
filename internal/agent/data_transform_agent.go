package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// DataTransformAgent performs simple string manipulations on the provided input.
type DataTransformAgent struct {
	agentID string
}

// NewDataTransformAgent creates a new DataTransformAgent.
func NewDataTransformAgent() *DataTransformAgent {
	return &DataTransformAgent{agentID: fmt.Sprintf("transform-agent-%s", uuid.NewString())}
}

// ID implements the Agent interface.
func (d *DataTransformAgent) ID() string { return d.agentID }

// Execute applies the requested operation to the input string.
// Supported operations: "uppercase", "lowercase", "reverse", "title".
// Defaults to "uppercase" when no operation is provided.
func (d *DataTransformAgent) Execute(ctx context.Context, task Task) Result {
	text, _ := task.Input["text"].(string)
	op, _ := task.Input["operation"].(string)
	if op == "" {
		op = "uppercase"
	}

	var out string
	switch op {
	case "uppercase":
		out = strings.ToUpper(text)
	case "lowercase":
		out = strings.ToLower(text)
	case "reverse":
		r := []rune(text)
		for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
		out = string(r)
	case "title":
		out = strings.Title(text)
	default:
		out = text
	}

	result := map[string]interface{}{
		"operation": op,
		"output":    out,
	}

	return Result{TaskID: task.ID, Output: result, Successful: true}
}

func init() {
	Register("DataTransformAgent", func() Agent { return NewDataTransformAgent() })
}
