package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"agentic.example.com/mvp/internal/agent"
)

// ExecutePipeline runs the provided pipeline sequentially while executing all
// steps within a group concurrently. Results from each step are stored in
// StepData so later steps can reference them.
func (o *Orchestrator) ExecutePipeline(ctx context.Context, p Pipeline, initialInput map[string]interface{}) (StepData, error) {
	fmt.Printf("Orchestrator: Starting pipeline '%s'\n", p.ID)

	current := make(StepData)
	for k, v := range initialInput {
		current[fmt.Sprintf("initial.%s", k)] = v
	}

	for gi, group := range p.Groups {
		fmt.Printf("Orchestrator: Executing group %d: '%s' with %d step(s)\n", gi+1, group.Name, len(group.Steps))

		var wg sync.WaitGroup
		type stepResult struct {
			step   PipelineStep
			result agent.Result
		}
		resultCh := make(chan stepResult, len(group.Steps))

		for _, st := range group.Steps {
			step := st
			wg.Add(1)
			go func() {
				defer wg.Done()

				taskInput := make(map[string]interface{})
				for target, source := range step.InputMappings {
					val, ok := resolveSourcePath(current, source)
					if !ok {
						fmt.Printf("Orchestrator: Warning - source '%s' not found for step '%s'\n", source, step.Name)
					}
					taskInput[target] = val
				}
				for k, v := range step.AgentConfig.Input {
					if _, exists := taskInput[k]; !exists {
						taskInput[k] = v
					}
				}

				ag, ok := agent.New(step.AgentType)
				if !ok {
					resultCh <- stepResult{step: step, result: agent.Result{TaskID: step.Name, Error: fmt.Errorf("unknown agent type '%s'", step.AgentType)}}
					return
				}

				// apply configuration for known agent types
				switch a := ag.(type) {
				case *agent.HTTPCallAgent:
					if method, ok := taskInput["method"].(string); ok {
						a.Method = method
					}
					if url, ok := taskInput["url"].(string); ok {
						a.URL = url
					}
					if hdrs, ok := taskInput["headers"].(map[string]string); ok {
						a.Headers = hdrs
					}
				}

				task := agent.Task{
					ID:          fmt.Sprintf("%s_task_for_%s", p.ID, step.Name),
					Description: step.AgentConfig.Description,
					Input:       taskInput,
				}

				res := ag.Execute(ctx, task)
				resultCh <- stepResult{step: step, result: res}
			}()
		}

		wg.Wait()
		close(resultCh)

		for res := range resultCh {
			if !res.result.Successful {
				return current, fmt.Errorf("step '%s' failed: %w", res.step.Name, res.result.Error)
			}

			fmt.Printf("Orchestrator: Step '%s' completed. Output: %v\n", res.step.Name, res.result.Output)

			if res.result.Output != nil {
				current[fmt.Sprintf("%s.%s", res.step.Name, DefaultOutputKey)] = res.result.Output
			}
			current[fmt.Sprintf("%s.task_id", res.step.Name)] = res.result.TaskID
			current[fmt.Sprintf("%s.successful", res.step.Name)] = res.result.Successful
		}
	}

	fmt.Printf("Orchestrator: Pipeline '%s' completed successfully.\n", p.ID)
	return current, nil
}

// resolveSourcePath retrieves a value from StepData based on a simple path.
func resolveSourcePath(data StepData, path string) (interface{}, bool) {
	val, ok := data[path]
	return val, ok
}
