package orchestrator

import (
	"context"
	"strings"
	"testing"

	"agentic.example.com/mvp/internal/agent"
)

func TestStreamingAgent(t *testing.T) {
	p := Pipeline{
		ID: "stream",
		Groups: []PipelineGroup{{
			Name: "emit",
			Steps: []PipelineStep{{
				Name:        "stream",
				AgentType:   "StreamingEchoAgent",
				AgentConfig: agent.Task{Input: map[string]interface{}{"text": "abc"}},
			}},
		}},
	}

	orc := NewOrchestrator()
	events, errCh := orc.RunPipeline(context.Background(), p, nil)
	var parts []string
	var final string
	for ev := range events {
		if ev.Partial {
			if s, ok := ev.Result.Output.(string); ok {
				parts = append(parts, s)
			}
		} else {
			if s, ok := ev.Result.Output.(string); ok {
				final = s
			}
		}
	}
	if err := <-errCh; err != nil {
		t.Fatalf("pipeline error: %v", err)
	}
	if strings.Join(parts, "") != "abc" {
		t.Fatalf("expected streamed tokens to form 'abc', got %q", strings.Join(parts, ""))
	}
	if final != "abc" {
		t.Fatalf("expected final output 'abc', got %q", final)
	}
}
