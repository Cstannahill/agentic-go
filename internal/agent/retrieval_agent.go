package agent

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"agentic.example.com/mvp/internal/tools"
	"agentic.example.com/mvp/internal/vectorstore"
)

// RetrievalAgent queries the vector store for similar documents.
type RetrievalAgent struct {
	id   string
	tool *tools.RetrievalTool
}

// NewRetrievalAgent creates a RetrievalAgent using the default store.
func NewRetrievalAgent() *RetrievalAgent {
	return &RetrievalAgent{
		id:   fmt.Sprintf("retrieval-agent-%s", uuid.NewString()),
		tool: tools.NewRetrievalTool(vectorstore.DefaultStore(), 5),
	}
}

// NewRetrievalAgentWithK allows configuring the number of documents to return.
func NewRetrievalAgentWithK(k int) *RetrievalAgent {
	return &RetrievalAgent{
		id:   fmt.Sprintf("retrieval-agent-%s", uuid.NewString()),
		tool: tools.NewRetrievalTool(vectorstore.DefaultStore(), k),
	}
}

func (r *RetrievalAgent) ID() string { return r.id }

func (r *RetrievalAgent) Execute(ctx context.Context, task Task) Result {
	out, err := r.tool.Run(ctx, task.Input)
	if err != nil {
		return Result{TaskID: task.ID, Error: err}
	}
	return Result{TaskID: task.ID, Output: out, Successful: true}
}

func init() {
	Register("RetrievalAgent", func() Agent { return NewRetrievalAgent() })
}
