package agent

import (
	"context"
	"fmt"

	"agentic.example.com/mvp/internal/workspace"
)

// WorkspaceAgent allows pipeline steps to write to or read from the shared workspace.
// The action is determined by the "mode" field which must be either "write" or "read".
// For write mode the task input must include "key" and "data" (string).
// For read mode the agent retrieves the value for "key" and returns it under "data".
type WorkspaceAgent struct {
	id string
}

func NewWorkspaceAgent() *WorkspaceAgent {
	return &WorkspaceAgent{id: "workspace-agent"}
}

func (w *WorkspaceAgent) ID() string { return w.id }

func (w *WorkspaceAgent) Execute(ctx context.Context, task Task) Result {
	mode, _ := task.Input["mode"].(string)
	key, _ := task.Input["key"].(string)

	switch mode {
	case "write":
		data, _ := task.Input["data"].(string)
		workspace.DefaultStore.Put(key, []byte(data))
		return Result{TaskID: task.ID, Output: map[string]interface{}{"written": key}, Successful: true}
	case "read":
		val, ok := workspace.DefaultStore.Get(key)
		if !ok {
			return Result{TaskID: task.ID, Error: fmt.Errorf("key not found"), Successful: false}
		}
		return Result{TaskID: task.ID, Output: map[string]interface{}{"data": string(val)}, Successful: true}
	default:
		return Result{TaskID: task.ID, Error: fmt.Errorf("unknown mode '%s'", mode), Successful: false}
	}
}

func init() {
	Register("WorkspaceAgent", func() Agent { return NewWorkspaceAgent() })
}
