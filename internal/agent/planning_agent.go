package agent

import (
	"context"
	"fmt"
	"github.com/google/uuid"
)

// PlannedStep mirrors the orchestrator PipelineStep structure but lives in the
// agent package to avoid import cycles.
type PlannedStep struct {
	Name          string
	AgentType     string
	AgentConfig   Task
	InputMappings map[string]string
	BranchKey     string
}

// PlanningAgent is an agent that converts a high level goal into a slice of
// planned steps which can later be converted to a pipeline definition.
type PlanningAgent interface {
	Agent
	Plan(goal string) ([]PlannedStep, error)
}

// SimplePlanningAgent is a trivial implementation used for examples and tests.
// It generates a single EchoAgent step that echoes the provided goal.
type SimplePlanningAgent struct {
	agentID string
}

func NewSimplePlanningAgent() *SimplePlanningAgent {
	return &SimplePlanningAgent{agentID: fmt.Sprintf("planner-%s", uuid.NewString())}
}

func (p *SimplePlanningAgent) ID() string { return p.agentID }

func (p *SimplePlanningAgent) Plan(goal string) ([]PlannedStep, error) {
	step := PlannedStep{
		Name:      "echo_goal",
		AgentType: "EchoAgent",
		AgentConfig: Task{
			Description: "echo planned goal",
			Input:       map[string]interface{}{"message": goal},
		},
		InputMappings: map[string]string{},
	}
	return []PlannedStep{step}, nil
}

// Execute satisfies the Agent interface. It expects the task input to contain a
// "goal" string and returns the planned steps as the result output.
func (p *SimplePlanningAgent) Execute(ctx context.Context, task Task) Result {
	goal, ok := task.Input["goal"].(string)
	if !ok {
		return Result{TaskID: task.ID, Error: fmt.Errorf("missing goal"), Successful: false}
	}
	steps, err := p.Plan(goal)
	if err != nil {
		return Result{TaskID: task.ID, Error: err, Successful: false}
	}
	return Result{TaskID: task.ID, Output: steps, Successful: true}
}

func init() {
	Register("SimplePlanningAgent", func() Agent { return NewSimplePlanningAgent() })
}
