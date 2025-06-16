package orchestrator

import (
	"context"
	"testing"

	"agentic.example.com/mvp/internal/agent"
)

type textAgent struct{ text string }

func (t *textAgent) ID() string { return "text-agent" }
func (t *textAgent) Execute(ctx context.Context, task agent.Task) agent.Result {
	out := t.text
	if val, ok := task.Input["text"].(string); ok {
		out = val
	}
	return agent.Result{TaskID: task.ID, Output: map[string]interface{}{"output": out}, Successful: true}
}

func init() {
	agent.Register("TextAgent", func() agent.Agent { return &textAgent{text: "bad"} })
}

func TestCriticRetry(t *testing.T) {
	p := Pipeline{
		ID: "critic",
		Groups: []PipelineGroup{{
			Name: "work",
			Steps: []PipelineStep{{
				Name:        "emit",
				AgentType:   "TextAgent",
				CriticType:  "KeywordCriticAgent",
				MaxRetries:  1,
				AgentConfig: agent.Task{Description: "emit"},
			}},
		}},
	}

	orc := NewOrchestrator()
	data, err := orc.ExecutePipeline(context.Background(), p, nil)
	if err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	out, _ := data["emit.default_output"].(map[string]interface{})
	if out["output"] != "good" {
		t.Fatalf("expected retry to adjust output to 'good', got %#v", out)
	}
}
