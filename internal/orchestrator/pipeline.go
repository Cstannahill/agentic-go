package orchestrator

import (
	"context"

	"agentic.example.com/mvp/internal/agent"
)

// StepOutputKey defines keys for outputs stored in pipeline state.
type StepOutputKey string

const (
	// DefaultOutputKey is used when a step does not specify a named output.
	DefaultOutputKey StepOutputKey = "default_output"
)

// StepData contains all data available to a pipeline step, including
// initial inputs and outputs from previous steps.
type StepData map[string]interface{}

// PipelineStep represents a single step in a pipeline.
type PipelineStep struct {
	Name          string            // Unique name for this step
	AgentType     string            // Identifies which agent implementation to use
	AgentConfig   agent.Task        // Base task configuration for the agent
	InputMappings map[string]string // Mapping of agent input keys to StepData sources
}

// Pipeline defines a sequence of steps executed by the orchestrator.
type Pipeline struct {
	ID          string
	Description string
	Steps       []PipelineStep
}

// Orchestrator coordinates execution of pipelines.
type Orchestrator struct {
	// Future: agent registry, logging, etc.
}

// NewOrchestrator returns a new orchestrator instance.
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{}
}

// ExecutableAgent mirrors the agent.Agent interface so orchestrator only
// relies on required methods.
type ExecutableAgent interface {
	ID() string
	Execute(ctx context.Context, task agent.Task) agent.Result
}
