package orchestrator

import (
	"context"
	"fmt"

	"agentic.example.com/mvp/internal/agent"
)

// ExecutePipeline runs the provided pipeline sequentially.
// Each step's output is stored in StepData and may be referenced by later steps.
func (o *Orchestrator) ExecutePipeline(ctx context.Context, p Pipeline, initialInput map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("Orchestrator: Starting pipeline '%s'\n", p.ID)

	// Initialize step data with initial inputs (prefixed with "initial.").
	current := make(StepData)
	for k, v := range initialInput {
		current[fmt.Sprintf("initial.%s", k)] = v
	}

	var final map[string]interface{}

	for i, step := range p.Steps {
		fmt.Printf("Orchestrator: Executing step %d: '%s' (AgentType: %s)\n", i+1, step.Name, step.AgentType)

		// Build task input from mappings.
		taskInput := make(map[string]interface{})
		for target, source := range step.InputMappings {
			val, ok := resolveSourcePath(current, source)
			if !ok {
				fmt.Printf("Orchestrator: Warning - source '%s' not found for step '%s'\n", source, step.Name)
			}
			taskInput[target] = val
		}
		// Merge static config values if not overridden by mappings.
		for k, v := range step.AgentConfig.Input {
			if _, exists := taskInput[k]; !exists {
				taskInput[k] = v
			}
		}

		// Instantiate agent via registry for plug-and-play behavior.
		agIntf, ok := agent.New(step.AgentType)
		if !ok {
			return nil, fmt.Errorf("unknown agent type '%s'", step.AgentType)
		}
		var ag ExecutableAgent = agIntf

		task := agent.Task{
			ID:          fmt.Sprintf("%s_task_for_%s", p.ID, step.Name),
			Description: step.AgentConfig.Description,
			Input:       taskInput,
		}

		// Run the agent in a goroutine and wait via channel for result.
		resultCh := make(chan agent.Result, 1)
		go func() { resultCh <- ag.Execute(ctx, task) }()

		result := <-resultCh
		if !result.Successful {
			return nil, fmt.Errorf("step '%s' failed: %w", step.Name, result.Error)
		}

		fmt.Printf("Orchestrator: Step '%s' completed. Output: %v\n", step.Name, result.Output)

		// Store output for later steps.
		if result.Output != nil {
			current[fmt.Sprintf("%s.%s", step.Name, DefaultOutputKey)] = result.Output
		}
		current[fmt.Sprintf("%s.task_id", step.Name)] = result.TaskID
		current[fmt.Sprintf("%s.successful", step.Name)] = result.Successful

		if outputMap, ok := result.Output.(map[string]interface{}); ok {
			final = outputMap
		} else {
			final = map[string]interface{}{string(DefaultOutputKey): result.Output}
		}
	}

	fmt.Printf("Orchestrator: Pipeline '%s' completed successfully.\n", p.ID)
	return final, nil
}

// resolveSourcePath retrieves a value from StepData based on a simple path.
func resolveSourcePath(data StepData, path string) (interface{}, bool) {
	val, ok := data[path]
	return val, ok
}
