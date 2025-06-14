package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"agentic.example.com/mvp/internal/agent"
)

// StepEvent represents the completion of a pipeline step. It is sent on a
// channel when using RunPipeline for asynchronous execution.
type StepEvent struct {
	Group  string
	Step   string
	Result agent.Result
}

// RunPipeline executes the pipeline and streams step results over a channel.
// An error channel is returned which will receive a single error if the
// pipeline fails. The returned channels are closed when execution completes.
// Step results are sent as soon as each step finishes.
func (o *Orchestrator) RunPipeline(ctx context.Context, p Pipeline, initialInput map[string]interface{}) (<-chan StepEvent, <-chan error) {
	events := make(chan StepEvent)
	errCh := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errCh)

		current := make(StepData)
		for k, v := range initialInput {
			current[fmt.Sprintf("initial.%s", k)] = v
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for gi, group := range p.Groups {
			fmt.Printf("Orchestrator: Executing group %d: '%s' with %d step(s)\n", gi+1, group.Name, len(group.Steps))

			var wg sync.WaitGroup
			resultCh := make(chan StepEvent, len(group.Steps))

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
						resultCh <- StepEvent{Group: group.Name, Step: step.Name, Result: agent.Result{TaskID: step.Name, Error: fmt.Errorf("unknown agent type '%s'", step.AgentType)}}
						return
					}

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
					resultCh <- StepEvent{Group: group.Name, Step: step.Name, Result: res}
				}()
			}

			wg.Wait()
			close(resultCh)

			for ev := range resultCh {
				if !ev.Result.Successful {
					errCh <- fmt.Errorf("step '%s' failed: %w", ev.Step, ev.Result.Error)
					cancel()
					return
				}

				fmt.Printf("Orchestrator: Step '%s' completed. Output: %v\n", ev.Step, ev.Result.Output)

				if ev.Result.Output != nil {
					current[fmt.Sprintf("%s.%s", ev.Step, DefaultOutputKey)] = ev.Result.Output
				}
				current[fmt.Sprintf("%s.task_id", ev.Step)] = ev.Result.TaskID
				current[fmt.Sprintf("%s.successful", ev.Step)] = ev.Result.Successful

				events <- ev
			}
		}

		errCh <- nil
	}()

	return events, errCh
}

// ExecutePipeline runs the provided pipeline sequentially while executing all
// steps within a group concurrently. Results from each step are stored in
// StepData so later steps can reference them.
func (o *Orchestrator) ExecutePipeline(ctx context.Context, p Pipeline, initialInput map[string]interface{}) (StepData, error) {
	fmt.Printf("Orchestrator: Starting pipeline '%s'\n", p.ID)

	current := make(StepData)
	for k, v := range initialInput {
		current[fmt.Sprintf("initial.%s", k)] = v
	}
	events, errCh := o.RunPipeline(ctx, p, initialInput)

	for ev := range events {
		if !ev.Result.Successful {
			// error will also be sent on errCh; wait for it
			continue
		}

		if ev.Result.Output != nil {
			current[fmt.Sprintf("%s.%s", ev.Step, DefaultOutputKey)] = ev.Result.Output
		}
		current[fmt.Sprintf("%s.task_id", ev.Step)] = ev.Result.TaskID
		current[fmt.Sprintf("%s.successful", ev.Step)] = ev.Result.Successful
	}

	if err := <-errCh; err != nil {
		return current, err
	}

	fmt.Printf("Orchestrator: Pipeline '%s' completed successfully.\n", p.ID)
	return current, nil
}

// resolveSourcePath retrieves a value from StepData based on a simple path.
func resolveSourcePath(data StepData, path string) (interface{}, bool) {
	// First try exact key match which preserves backwards compatibility.
	if val, ok := data[path]; ok {
		return val, true
	}

	// Support dotted paths for nested map lookups. A path like
	// "step.output.field" will attempt to resolve "step.output" from
	// StepData and then drill into the resulting map for "field".
	parts := strings.Split(path, ".")
	for i := len(parts) - 1; i > 0; i-- {
		key := strings.Join(parts[:i], ".")
		val, ok := data[key]
		if !ok {
			continue
		}
		cur := val
		found := true
		for _, p := range parts[i:] {
			m, ok := cur.(map[string]interface{})
			if !ok {
				found = false
				break
			}
			cur, ok = m[p]
			if !ok {
				found = false
				break
			}
		}
		if found {
			return cur, true
		}
	}
	return nil, false
}
