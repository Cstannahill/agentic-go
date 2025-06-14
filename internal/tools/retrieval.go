package tools

import (
	"context"
	"fmt"
)

// RetrievalTool is a stub that would query an index or database.
func RetrievalTool(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	query, _ := input["query"].(string)
	result := map[string]interface{}{
		"documents": []string{fmt.Sprintf("doc_for_%s", query)},
	}
	return result, nil
}
