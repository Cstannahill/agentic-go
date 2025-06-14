package tools

import "context"

// Tool defines an executable piece of functionality the orchestrator can call.
type Tool interface {
	Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}
