package orchestrator

import (
	"agentic.example.com/mvp/internal/agent"
)

// BuildRAGPipeline returns a preconfigured retrieval augmented generation pipeline.
// The pipeline expects initial input with keys:
//
//		query                - user query text
//		template             - prompt template string
//		model                - optional model name for generation
//		top_k                - optional number of documents to retrieve
//		completion_endpoint  - optional override for the generation endpoint
//	     extra_context        - optional map merged into the prompt data
func BuildRAGPipeline(id string) Pipeline {
	return Pipeline{
		ID:          id,
		Description: "retrieval augmented generation",
		Groups: []PipelineGroup{
			{
				Name: "embed",
				Steps: []PipelineStep{
					{
						Name:        "embed_query",
						AgentType:   "EmbeddingAgent",
						AgentConfig: agent.Task{Description: "Embed user query"},
						InputMappings: map[string]string{
							"text": "initial.query",
						},
					},
				},
			},
			{
				Name: "retrieve",
				Steps: []PipelineStep{
					{
						Name:        "retrieve_docs",
						AgentType:   "RetrievalAgent",
						AgentConfig: agent.Task{Description: "Retrieve documents"},
						InputMappings: map[string]string{
							"embedding": "embed_query.default_output.embedding",
							"top_k":     "initial.top_k",
						},
					},
				},
			},
			{
				Name: "rerank",
				Steps: []PipelineStep{
					{
						Name:        "rerank_docs",
						AgentType:   "RerankAgent",
						AgentConfig: agent.Task{Description: "Rerank documents"},
						InputMappings: map[string]string{
							"documents": "retrieve_docs.default_output.documents",
							"query":     "initial.query",
						},
					},
				},
			},
			{
				Name: "prompt",
				Steps: []PipelineStep{
					{
						Name:        "build_prompt",
						AgentType:   "PromptAgent",
						AgentConfig: agent.Task{Description: "Inject context"},
						InputMappings: map[string]string{
							"template":  "initial.template",
							"documents": "rerank_docs.default_output.reranked",
							"query":     "initial.query",
							"context":   "initial.extra_context",
						},
					},
				},
			},
			{
				Name: "generate",
				Steps: []PipelineStep{
					{
						Name:        "generate_answer",
						AgentType:   "GenerationAgent",
						AgentConfig: agent.Task{Description: "Generate final answer"},
						InputMappings: map[string]string{
							"prompt":   "build_prompt.default_output.prompt",
							"model":    "initial.model",
							"endpoint": "initial.completion_endpoint",
						},
					},
				},
			},
		},
	}
}

type ContextDocument struct {
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Score    float64                `json:"score,omitempty"`
}

type RAGResponse struct {
	Query     string            `json:"query"`
	Answer    string            `json:"answer"`
	Documents []ContextDocument `json:"documents"`
}

// ExtractRAGResponse builds a structured RAGResponse from StepData
// returned by ExecutePipeline on a pipeline produced by BuildRAGPipeline.
func ExtractRAGResponse(data StepData) (RAGResponse, bool) {
	genOut, ok := data["generate_answer.default_output"].(map[string]interface{})
	if !ok {
		return RAGResponse{}, false
	}
	rerankOut, ok := data["rerank_docs.default_output"].(map[string]interface{})
	if !ok {
		return RAGResponse{}, false
	}
	answer, _ := genOut["completion"].(string)
	query, _ := data["initial.query"].(string)
	ctx, _ := rerankOut["reranked"].([]map[string]interface{})
	docs := make([]ContextDocument, len(ctx))
	for i, d := range ctx {
		docs[i] = ContextDocument{}
		if id, ok := d["id"].(string); ok {
			docs[i].ID = id
		}
		if meta, ok := d["metadata"].(map[string]interface{}); ok {
			docs[i].Metadata = meta
		}
		if score, ok := d["score"].(float64); ok {
			docs[i].Score = score
		}
	}
	return RAGResponse{Query: query, Answer: answer, Documents: docs}, true
}
