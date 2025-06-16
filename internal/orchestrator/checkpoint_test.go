package orchestrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"agentic.example.com/mvp/internal/agent"
)

func TestCheckpointResume(t *testing.T) {
	tmpDir := t.TempDir()
	cp := NewCheckpointManager(tmpDir)

	p := Pipeline{
		ID: "resume",
		Groups: []PipelineGroup{
			{
				Name: "first",
				Steps: []PipelineStep{{
					Name:        "step1",
					AgentType:   "EchoAgent",
					AgentConfig: agent.Task{Input: map[string]interface{}{"message": "one"}},
				}},
			},
			{
				Name: "second",
				Steps: []PipelineStep{{
					Name:      "step2",
					AgentType: "EchoAgent",
					InputMappings: map[string]string{
						"message": "step1.default_output.processed_input.message",
					},
				}},
			},
		},
	}

	orc := NewOrchestrator()
	// Run first group only to simulate prior partial execution
	first := Pipeline{ID: p.ID, Groups: p.Groups[:1]}
	data, err := orc.ExecutePipeline(context.Background(), first, nil)
	if err != nil {
		t.Fatalf("first group run error: %v", err)
	}
	if err := cp.Save(p.ID, 1, data); err != nil {
		t.Fatalf("save checkpoint: %v", err)
	}

	// Resume and complete pipeline
	final, err := orc.ExecutePipelineWithCheckpoint(context.Background(), p, nil, cp)
	if err != nil {
		t.Fatalf("resume error: %v", err)
	}

	if _, ok := final["step2.default_output"]; !ok {
		t.Fatalf("expected step2 to run")
	}

	path := filepath.Join(tmpDir, p.ID+".json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("checkpoint should be removed")
	}
}
