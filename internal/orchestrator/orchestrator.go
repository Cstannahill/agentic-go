package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"agentic.example.com/mvp/internal/agent"
)

// ExecutePipeline runs the provided pipeline sequentially.
// Each step's output is stored in StepData and may be referenced by later steps.
func (o *Orchestrator) ExecutePipeline(ctx context.Context, p Pipeline, initialInput map[string]interface{}) (StepData, error) {
	fmt.Printf("Orchestrator: Starting pipeline '%s'\n", p.ID)

	// Initialize step data with initial inputs (prefixed with "initial.").
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

		for _, step := range group.Steps {
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

		}

		// Instantiate agent via registry for plug-and-play behavior.
		agIntf, ok := agent.New(step.AgentType)
		if !ok {
			return nil, fmt.Errorf("unknown agent type '%s'", step.AgentType)
		}
		var ag ExecutableAgent = agIntf

			var ag ExecutableAgent
			switch step.AgentType {
			case "EchoAgent":
				ag = agent.NewEchoAgent()
			case "HTTPCallAgent":
				method, _ := step.AgentConfig.Input["method"].(string)
				url, _ := step.AgentConfig.Input["url"].(string)
				ag = agent.NewHTTPCallAgent(method, url, map[string]string{})
			default:
				return current, fmt.Errorf("unknown agent type '%s'", step.AgentType)
			}

			task := agent.Task{
				ID:          fmt.Sprintf("%s_task_for_%s", p.ID, step.Name),
				Description: step.AgentConfig.Description,
				Input:       taskInput,
			}

			wg.Add(1)
			go func(st PipelineStep, ag ExecutableAgent, tk agent.Task) {
				defer wg.Done()
				result := ag.Execute(ctx, tk)
				resultCh <- stepResult{step: st, result: result}
			}(step, ag, task)
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
