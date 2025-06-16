package orchestrator

import (
	"context"
	"testing"

	"agentic.example.com/mvp/internal/agent"
)

type branchAgent struct{ label string }

func (b *branchAgent) ID() string { return "branch-agent" }

func (b *branchAgent) Execute(ctx context.Context, task agent.Task) agent.Result {
	return agent.Result{TaskID: task.ID, Output: map[string]interface{}{"branch": b.label}, Successful: true, Branch: b.label}
}

func init() {
	agent.Register("BranchTestAgent", func() agent.Agent { return &branchAgent{label: "foo"} })
}

func TestBranching(t *testing.T) {
	p := Pipeline{
		ID: "branch",
		Groups: []PipelineGroup{{
			Name:  "decide",
			Steps: []PipelineStep{{Name: "decide", AgentType: "BranchTestAgent", BranchKey: "branch"}},
		}},
		Branches: map[string]PipelineGroup{
			"foo": {
				Name:  "foo",
				Steps: []PipelineStep{{Name: "foo_step", AgentType: "EchoAgent"}},
			},
			"default": {
				Name:  "default",
				Steps: []PipelineStep{{Name: "default_step", AgentType: "EchoAgent"}},
			},
		},
	}
	orc := NewOrchestrator()
	ctx := context.Background()
	data, err := orc.ExecutePipeline(ctx, p, nil)
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if _, ok := data["foo_step.default_output"]; !ok {
		t.Fatalf("expected foo branch to run")
	}
	if _, ok := data["default_step.default_output"]; ok {
		t.Fatalf("default branch should not run")
	}
}
