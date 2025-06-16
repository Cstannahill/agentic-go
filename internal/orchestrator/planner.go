package orchestrator

import (
	"agentic.example.com/mvp/internal/agent"
	"context"
	"fmt"
	"github.com/google/uuid"
)

// ExecutePlanningPipeline obtains a plan from the provided PlanningAgent using
// the goal string and then executes the resulting pipeline. The planner output
// is converted into a single pipeline group.
func (o *Orchestrator) ExecutePlanningPipeline(ctx context.Context, planner agent.PlanningAgent, goal string) (StepData, error) {
	planRes := planner.Execute(ctx, agent.Task{
		ID:          fmt.Sprintf("plan-%s", uuid.NewString()),
		Description: "generate pipeline plan",
		Input:       map[string]interface{}{"goal": goal},
	})
	if !planRes.Successful {
		return nil, fmt.Errorf("planning failed: %w", planRes.Error)
	}

	steps, ok := planRes.Output.([]agent.PlannedStep)
	if !ok {
		return nil, fmt.Errorf("planner returned unexpected output")
	}

	pipelineSteps := make([]PipelineStep, len(steps))
	for i, st := range steps {
		pipelineSteps[i] = PipelineStep{
			Name:          st.Name,
			AgentType:     st.AgentType,
			AgentConfig:   st.AgentConfig,
			InputMappings: st.InputMappings,
			BranchKey:     st.BranchKey,
		}
	}

	p := Pipeline{
		ID:     fmt.Sprintf("dynamic-%s", uuid.NewString()),
		Groups: []PipelineGroup{{Name: "generated", Steps: pipelineSteps}},
	}

	return o.ExecutePipeline(ctx, p, nil)
}
