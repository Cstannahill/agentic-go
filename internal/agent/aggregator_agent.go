package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// AggregatorAgent selects the best branch result among candidates.
type AggregatorAgent interface {
	Choose(ctx context.Context, branches map[string]StepData) (string, StepData)
	Agent
}

// StepData mirrors orchestrator.StepData for inter-package use.
type StepData map[string]interface{}

// LengthAggregator picks the branch whose output string is longest.
// It expects each branch's StepData to contain a key ending with
// `.message` whose value is a string.
type LengthAggregator struct {
	id string
}

func NewLengthAggregator() *LengthAggregator {
	return &LengthAggregator{id: fmt.Sprintf("agg-%s", uuid.NewString())}
}

func (l *LengthAggregator) ID() string { return l.id }

func (l *LengthAggregator) Execute(ctx context.Context, task Task) Result {
	branches, ok := task.Input["branches"].(map[string]StepData)
	if !ok {
		return Result{TaskID: task.ID, Error: fmt.Errorf("missing branches")}
	}
	label, data := l.Choose(ctx, branches)
	return Result{TaskID: task.ID, Output: map[string]interface{}{"label": label, "data": data}, Successful: true}
}

func (l *LengthAggregator) Choose(ctx context.Context, branches map[string]StepData) (string, StepData) {
	best := ""
	var sel StepData
	max := -1
	for lbl, data := range branches {
		for _, v := range data {
			if m, ok := v.(map[string]interface{}); ok {
				if s, ok := m["message"].(string); ok {
					if len(s) > max {
						max = len(s)
						best = lbl
						sel = data
					}
				}
			}
		}
	}
	return best, sel
}

func init() {
	Register("LengthAggregator", func() Agent { return NewLengthAggregator() })
}
