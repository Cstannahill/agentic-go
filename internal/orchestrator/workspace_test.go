package orchestrator

import (
	"context"
	"testing"

	"agentic.example.com/mvp/internal/agent"
)

func TestWorkspaceSharing(t *testing.T) {
	p := Pipeline{
		ID: "workspace",
		Groups: []PipelineGroup{
			{
				Name: "write",
				Steps: []PipelineStep{{
					Name:        "store",
					AgentType:   "WorkspaceAgent",
					AgentConfig: agent.Task{Input: map[string]interface{}{"mode": "write", "key": "foo", "data": "bar"}},
				}},
			},
			{
				Name: "read",
				Steps: []PipelineStep{{
					Name:        "load",
					AgentType:   "WorkspaceAgent",
					AgentConfig: agent.Task{Input: map[string]interface{}{"mode": "read", "key": "foo"}},
				}},
			},
		},
	}

	orc := NewOrchestrator()
	data, err := orc.ExecutePipeline(context.Background(), p, nil)
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	out, _ := data["load.default_output"].(map[string]interface{})
	if out["data"] != "bar" {
		t.Fatalf("expected to load 'bar' from workspace, got %#v", out)
	}
}
