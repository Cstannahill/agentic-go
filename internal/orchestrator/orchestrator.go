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

		groups := p.Groups
		for gi := 0; gi < len(groups); gi++ {
			group := groups[gi]
			var groupBranch string
			fmt.Printf("Orchestrator: Executing group %d: '%s' with %d step(s)\n", gi+1, group.Name, len(group.Steps))

			var wg sync.WaitGroup
			resultCh := make(chan StepEvent, len(group.Steps))
			stepMap := make(map[string]PipelineStep)

			for _, st := range group.Steps {
				step := st
				stepMap[step.Name] = step
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

					var res agent.Result
					for attempt := 0; ; attempt++ {
						res = ag.Execute(ctx, task)

						// Run critic if configured
						if step.CriticType != "" {
							cagt, ok := agent.New(step.CriticType)
							if !ok {
								res = agent.Result{TaskID: task.ID, Error: fmt.Errorf("unknown critic type '%s'", step.CriticType)}
								break
							}
							critic, ok := cagt.(agent.CriticAgent)
							if !ok {
								res = agent.Result{TaskID: task.ID, Error: fmt.Errorf("critic agent '%s' missing interface", step.CriticType)}
								break
							}
							fb := critic.Review(ctx, res)
							if !fb.Approved {
								if fb.Retry && attempt < step.MaxRetries {
									if fb.AdjustedInput != nil {
										for k, v := range fb.AdjustedInput {
											task.Input[k] = v
										}
									}
									continue
								}
								res.Successful = false
								if fb.Escalate {
									res.Error = fmt.Errorf("critic escalation for step '%s'", step.Name)
								} else {
									res.Error = fmt.Errorf("critic rejected step '%s'", step.Name)
								}
							}
						}
						break
					}

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

				if st, ok := stepMap[ev.Step]; ok && st.BranchKey != "" {
					if ev.Result.Branch != "" {
						groupBranch = ev.Result.Branch
					} else if outMap, ok := ev.Result.Output.(map[string]interface{}); ok {
						if lbl, ok := outMap[st.BranchKey].(string); ok {
							groupBranch = lbl
						}
					}
				}

				events <- ev
			}

			if groupBranch != "" {
				next, ok := p.Branches[groupBranch]
				if !ok {
					next, ok = p.Branches["default"]
				}
				if ok && len(next) > 0 {
					if len(next) == 1 {
						groups = append(groups[:gi+1], append([]PipelineGroup{next[0]}, groups[gi+1:]...)...)
					} else {
						branchResults := make(map[string]agent.StepData)
						for i, grp := range next {
							out, err := o.runSimple(ctx, []PipelineGroup{grp}, current)
							if err != nil {
								errCh <- err
								cancel()
								return
							}
							label := fmt.Sprintf("%s_%d", groupBranch, i)
							branchResults[label] = agent.StepData(out)
						}
						if p.AggregatorType != "" {
							ag, ok := agent.New(p.AggregatorType)
							if ok {
								if agg, ok := ag.(agent.AggregatorAgent); ok {
									_, data := agg.Choose(ctx, branchResults)
									for k, v := range data {
										current[k] = v
									}
								}
							}
						}
					}
				}
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

// ExecutePipelineWithCheckpoint behaves like ExecutePipeline but persists
// progress after each group using the provided CheckpointManager. If a
// checkpoint exists for the pipeline it resumes from the saved group index.
func (o *Orchestrator) ExecutePipelineWithCheckpoint(ctx context.Context, p Pipeline, initialInput map[string]interface{}, cp *CheckpointManager) (StepData, error) {
	fmt.Printf("Orchestrator: Starting pipeline '%s' with checkpoints\n", p.ID)

	start := 0
	current := make(StepData)
	for k, v := range initialInput {
		current[fmt.Sprintf("initial.%s", k)] = v
	}
	if idx, data, err := cp.Load(p.ID); err == nil {
		start = idx
		for k, v := range data {
			current[k] = v
		}
	}

	groups := p.Groups
	for gi := start; gi < len(groups); gi++ {
		group := groups[gi]
		var groupBranch string
		fmt.Printf("Orchestrator: Executing group %d: '%s' with %d step(s)\n", gi+1, group.Name, len(group.Steps))

		var wg sync.WaitGroup
		resultCh := make(chan StepEvent, len(group.Steps))
		stepMap := make(map[string]PipelineStep)

		for _, st := range group.Steps {
			step := st
			stepMap[step.Name] = step
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

				task := agent.Task{
					ID:          fmt.Sprintf("%s_task_for_%s", p.ID, step.Name),
					Description: step.AgentConfig.Description,
					Input:       taskInput,
				}

				res := ag.Execute(ctx, task)

				// Critic logic reused from RunPipeline
				if step.CriticType != "" {
					cagt, ok := agent.New(step.CriticType)
					if !ok {
						res = agent.Result{TaskID: task.ID, Error: fmt.Errorf("unknown critic type '%s'", step.CriticType)}
					} else if critic, ok := cagt.(agent.CriticAgent); ok {
						for attempt := 0; ; attempt++ {
							fb := critic.Review(ctx, res)
							if fb.Approved {
								break
							}
							if fb.Retry && attempt < step.MaxRetries {
								if fb.AdjustedInput != nil {
									for k, v := range fb.AdjustedInput {
										task.Input[k] = v
									}
								}
								res = ag.Execute(ctx, task)
								continue
							}
							res.Successful = false
							if fb.Escalate {
								res.Error = fmt.Errorf("critic escalation for step '%s'", step.Name)
							} else {
								res.Error = fmt.Errorf("critic rejected step '%s'", step.Name)
							}
							break
						}
					} else {
						res = agent.Result{TaskID: task.ID, Error: fmt.Errorf("critic agent '%s' missing interface", step.CriticType)}
					}
				}

				resultCh <- StepEvent{Group: group.Name, Step: step.Name, Result: res}
			}()
		}

		wg.Wait()
		close(resultCh)

		for ev := range resultCh {
			if !ev.Result.Successful {
				cp.Save(p.ID, gi, current)
				return current, fmt.Errorf("step '%s' failed: %w", ev.Step, ev.Result.Error)
			}

			if ev.Result.Output != nil {
				current[fmt.Sprintf("%s.%s", ev.Step, DefaultOutputKey)] = ev.Result.Output
			}
			current[fmt.Sprintf("%s.task_id", ev.Step)] = ev.Result.TaskID
			current[fmt.Sprintf("%s.successful", ev.Step)] = ev.Result.Successful

			if st, ok := stepMap[ev.Step]; ok && st.BranchKey != "" {
				if ev.Result.Branch != "" {
					groupBranch = ev.Result.Branch
				} else if outMap, ok := ev.Result.Output.(map[string]interface{}); ok {
					if lbl, ok := outMap[st.BranchKey].(string); ok {
						groupBranch = lbl
					}
				}
			}
		}

		if err := cp.Save(p.ID, gi+1, current); err != nil {
			return current, err
		}

		if groupBranch != "" {
			next, ok := p.Branches[groupBranch]
			if !ok {
				next, ok = p.Branches["default"]
			}
			if ok && len(next) > 0 {
				if len(next) == 1 {
					groups = append(groups[:gi+1], append([]PipelineGroup{next[0]}, groups[gi+1:]...)...)
				} else {
					branchResults := make(map[string]agent.StepData)
					for i, grp := range next {
						out, err := o.runSimple(ctx, []PipelineGroup{grp}, current)
						if err != nil {
							cp.Save(p.ID, gi, current)
							return current, err
						}
						label := fmt.Sprintf("%s_%d", groupBranch, i)
						branchResults[label] = agent.StepData(out)
					}
					if p.AggregatorType != "" {
						ag, ok := agent.New(p.AggregatorType)
						if ok {
							if agg, ok := ag.(agent.AggregatorAgent); ok {
								_, data := agg.Choose(ctx, branchResults)
								for k, v := range data {
									current[k] = v
								}
							}
						}
					}
				}
			}
		}
	}

	cp.Remove(p.ID)
	fmt.Printf("Orchestrator: Pipeline '%s' completed successfully.\n", p.ID)
	return current, nil
}

// runSimple executes the provided groups sequentially starting from the given
// step data. It is used for speculative branches where full pipeline features
// like checkpointing are unnecessary.
func (o *Orchestrator) runSimple(ctx context.Context, groups []PipelineGroup, start StepData) (StepData, error) {
	current := make(StepData)
	for k, v := range start {
		current[k] = v
	}
	for _, group := range groups {
		var wg sync.WaitGroup
		resultCh := make(chan StepEvent, len(group.Steps))
		for _, st := range group.Steps {
			step := st
			wg.Add(1)
			go func() {
				defer wg.Done()
				taskInput := make(map[string]interface{})
				for target, source := range step.InputMappings {
					val, _ := resolveSourcePath(current, source)
					taskInput[target] = val
				}
				for k, v := range step.AgentConfig.Input {
					if _, ok := taskInput[k]; !ok {
						taskInput[k] = v
					}
				}
				ag, ok := agent.New(step.AgentType)
				if !ok {
					resultCh <- StepEvent{Group: group.Name, Step: step.Name, Result: agent.Result{TaskID: step.Name, Error: fmt.Errorf("unknown agent type '%s'", step.AgentType)}}
					return
				}
				task := agent.Task{ID: fmt.Sprintf("%s_task", step.Name), Description: step.AgentConfig.Description, Input: taskInput}
				res := ag.Execute(ctx, task)
				resultCh <- StepEvent{Group: group.Name, Step: step.Name, Result: res}
			}()
		}
		wg.Wait()
		close(resultCh)
		for ev := range resultCh {
			if !ev.Result.Successful {
				return current, fmt.Errorf("step '%s' failed: %w", ev.Step, ev.Result.Error)
			}
			if ev.Result.Output != nil {
				current[fmt.Sprintf("%s.%s", ev.Step, DefaultOutputKey)] = ev.Result.Output
			}
			current[fmt.Sprintf("%s.task_id", ev.Step)] = ev.Result.TaskID
			current[fmt.Sprintf("%s.successful", ev.Step)] = ev.Result.Successful
		}
	}
	return current, nil
}
