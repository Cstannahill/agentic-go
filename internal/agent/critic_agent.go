package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// CriticResult captures the outcome of a critic review.
type CriticResult struct {
	Approved      bool
	Retry         bool
	AdjustedInput map[string]interface{}
	Escalate      bool
}

// CriticAgent reviews the output of another agent.
type CriticAgent interface {
	Review(ctx context.Context, result Result) CriticResult
}

// KeywordCriticAgent requests a retry when the worker output contains a
// forbidden word. It optionally suggests a replacement text for the retry.
type KeywordCriticAgent struct {
	id         string
	RejectWord string
	Suggest    string
}

// NewKeywordCriticAgent creates a critic that rejects the given word.
func NewKeywordCriticAgent(reject, suggest string) *KeywordCriticAgent {
	return &KeywordCriticAgent{
		id:         fmt.Sprintf("critic-%s", uuid.NewString()),
		RejectWord: reject,
		Suggest:    suggest,
	}
}

func (k *KeywordCriticAgent) ID() string { return k.id }

// Execute satisfies the Agent interface for compatibility with the registry.
// The task input must include a `result` field containing the worker Result.
func (k *KeywordCriticAgent) Execute(ctx context.Context, task Task) Result {
	res, ok := task.Input["result"].(Result)
	if !ok {
		return Result{TaskID: task.ID, Error: fmt.Errorf("missing result")}
	}
	fb := k.Review(ctx, res)
	return Result{TaskID: task.ID, Output: fb, Successful: true}
}

// Review implements the CriticAgent interface.
func (k *KeywordCriticAgent) Review(ctx context.Context, res Result) CriticResult {
	var text string
	if m, ok := res.Output.(map[string]interface{}); ok {
		if s, ok := m["output"].(string); ok {
			text = s
		} else if s, ok := m["message"].(string); ok {
			text = s
		}
	} else if s, ok := res.Output.(string); ok {
		text = s
	}

	if strings.Contains(strings.ToLower(text), strings.ToLower(k.RejectWord)) {
		fb := CriticResult{Approved: false, Retry: true}
		if k.Suggest != "" {
			fb.AdjustedInput = map[string]interface{}{"text": k.Suggest}
		}
		return fb
	}
	return CriticResult{Approved: true}
}

func init() {
	Register("KeywordCriticAgent", func() Agent { return NewKeywordCriticAgent("bad", "good") })
}
