package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ContextBuilderAgent converts retrieved documents into a string of context
// that can be injected into downstream prompts. It joins the selected field
// from each document using the provided separator and merges any extra context
// values into the output map.
type ContextBuilderAgent struct {
	id string
}

// NewContextBuilderAgent creates a ContextBuilderAgent with a unique ID.
func NewContextBuilderAgent() *ContextBuilderAgent {
	return &ContextBuilderAgent{id: fmt.Sprintf("ctx-builder-%s", uuid.NewString())}
}

func (c *ContextBuilderAgent) ID() string { return c.id }

func getNested(m map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	cur := interface{}(m)
	for _, p := range parts {
		mp, ok := cur.(map[string]interface{})
		if !ok {
			return nil, false
		}
		cur, ok = mp[p]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}

// Execute expects input keys:
//
//	documents  - slice of document maps from retrieval
//	field      - optional dotted path within each document to extract (defaults to "metadata.text")
//	separator  - optional join string used between documents (defaults to "\n")
//	max_chars  - optional integer limiting the length of the combined context
//	extra      - optional map merged into the output context
//
// The output is a map[string]interface{} that can be provided directly to PromptAgent
// under its "context" input. When max_chars is set and the context is truncated,
// a boolean `truncated` field will also be present.
func (c *ContextBuilderAgent) Execute(ctx context.Context, task Task) Result {
	docs, ok := task.Input["documents"].([]map[string]interface{})
	if !ok {
		return Result{TaskID: task.ID, Error: fmt.Errorf("documents required")}
	}
	field, _ := task.Input["field"].(string)
	if field == "" {
		field = "metadata.text"
	}
	sep, _ := task.Input["separator"].(string)
	if sep == "" {
		sep = "\n"
	}
	maxChars, _ := task.Input["max_chars"].(int)

	var parts []string
	for _, d := range docs {
		if val, ok := getNested(d, field); ok {
			if s, ok := val.(string); ok && s != "" {
				parts = append(parts, s)
			}
		}
	}
	combined := strings.Join(parts, sep)
	truncated := false
	if maxChars > 0 && len(combined) > maxChars {
		combined = combined[:maxChars]
		truncated = true
	}
	contextMap := map[string]interface{}{"retrieved_context": combined}
	if truncated {
		contextMap["truncated"] = true
	}
	if extra, ok := task.Input["extra"].(map[string]interface{}); ok {
		for k, v := range extra {
			contextMap[k] = v
		}
	}
	return Result{TaskID: task.ID, Output: contextMap, Successful: true}
}

func init() {
	Register("ContextBuilderAgent", func() Agent { return NewContextBuilderAgent() })
}
