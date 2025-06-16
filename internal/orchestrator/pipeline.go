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
	// BranchKey, when set, specifies the result field containing a branch
	// label. The orchestrator uses this label to select the next pipeline
	// group.
	BranchKey    string
	CriticType   string     // Optional critic agent for feedback
	CriticConfig agent.Task // Base task for the critic agent
	MaxRetries   int        // Number of times to retry on critic request
}

// Pipeline defines a sequence of steps executed by the orchestrator.
// PipelineGroup represents a set of steps that may be executed concurrently.
// Steps inside a group should not depend on each other's output. Groups are
// executed sequentially in the order they are defined.
type PipelineGroup struct {
	Name  string
	Steps []PipelineStep
}

// Pipeline defines a sequence of groups executed by the orchestrator.  The
// design allows future expansion to parallel step execution while keeping the
// definition simple.
type Pipeline struct {
	ID          string
	Description string
	Groups      []PipelineGroup
	// Branches maps branch labels to one or more groups executed when a
	// step emits the corresponding label. A "default" entry may be provided.
	Branches map[string][]PipelineGroup
	// AggregatorType optionally specifies an agent used to select the best
	// result when multiple branch groups are executed speculatively.
	AggregatorType   string
	AggregatorConfig agent.Task
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
