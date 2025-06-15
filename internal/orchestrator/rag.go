package orchestrator

import (
	"agentic.example.com/mvp/internal/agent"
)

// RAGPipelineOptions customize construction of the retrieval augmented
// generation pipeline. Zero values fall back to defaults provided by the
// individual tools and agents.
type RAGPipelineOptions struct {
	// DefaultTopK controls how many documents are retrieved when the caller
	// does not provide a specific value in the initial input.
	DefaultTopK int

	// DefaultCompletionEndpoint specifies the HTTP endpoint used by the
	// GenerationAgent when no override is supplied at runtime.
	DefaultCompletionEndpoint string

	// DefaultContextField specifies which field the ContextBuilderAgent
	// extracts from each document when not provided by the caller.
	DefaultContextField string

	// DefaultSeparator defines the string used to join documents in the
	// retrieved context when not supplied in the initial input.
	DefaultSeparator string

	// DefaultMaxContextChars limits the length of the combined context when
	// the caller does not specify `max_chars`.
	DefaultMaxContextChars int

	// EnableReasoning adds an additional reasoning step using a second
	// GenerationAgent call. The template must be supplied at runtime under
	// the key `reason_template`.
	EnableReasoning bool
}

// BuildRAGPipeline returns a preconfigured retrieval augmented generation pipeline.
// Options provide defaults for retrieval depth and completion endpoint when
// the initial input omits them. At minimum a `query` and prompt `template`
// must be supplied.
func BuildRAGPipeline(id string, opts RAGPipelineOptions) Pipeline {
	groups := []PipelineGroup{
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
					Name:      "retrieve_docs",
					AgentType: "RetrievalAgent",
					AgentConfig: agent.Task{
						Description: "Retrieve documents",
						Input:       map[string]interface{}{"top_k": opts.DefaultTopK},
					},
					InputMappings: map[string]string{
						"embedding": "embed_query.default_output.embedding",
						"top_k":     "initial.top_k",
						"filter":    "initial.filter",
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
			Name: "context",
			Steps: []PipelineStep{
				{
					Name:      "build_context",
					AgentType: "ContextBuilderAgent",
					AgentConfig: agent.Task{
						Description: "Format retrieved context",
						Input: map[string]interface{}{
							"field":     opts.DefaultContextField,
							"separator": opts.DefaultSeparator,
							"max_chars": opts.DefaultMaxContextChars,
						},
					},
					InputMappings: map[string]string{
						"documents": "rerank_docs.default_output.reranked",
						"extra":     "initial.extra_context",
						"field":     "initial.context_field",
						"separator": "initial.context_separator",
						"max_chars": "initial.context_max_chars",
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
						"context":   "build_context.default_output",
					},
				},
			},
		},
		{
			Name: "generate",
			Steps: []PipelineStep{
				{
					Name:      "generate_answer",
					AgentType: "GenerationAgent",
					AgentConfig: agent.Task{
						Description: "Generate final answer",
						Input:       map[string]interface{}{"endpoint": opts.DefaultCompletionEndpoint},
					},
					InputMappings: map[string]string{
						"prompt":   "build_prompt.default_output.prompt",
						"model":    "initial.model",
						"endpoint": "initial.completion_endpoint",
					},
				},
			},
		},
	}

	if opts.EnableReasoning {
		groups = append(groups, PipelineGroup{
			Name: "reason",
			Steps: []PipelineStep{
				{
					Name:        "build_reason_prompt",
					AgentType:   "PromptAgent",
					AgentConfig: agent.Task{Description: "Build reasoning prompt"},
					InputMappings: map[string]string{
						"template":  "initial.reason_template",
						"documents": "rerank_docs.default_output.reranked",
						"query":     "initial.query",
						"answer":    "generate_answer.default_output.completion",
						"context":   "build_context.default_output",
					},
				},
				{
					Name:      "generate_reasoning",
					AgentType: "GenerationAgent",
					AgentConfig: agent.Task{
						Description: "Generate reasoning",
						Input:       map[string]interface{}{"endpoint": opts.DefaultCompletionEndpoint},
					},
					InputMappings: map[string]string{
						"prompt":   "build_reason_prompt.default_output.prompt",
						"model":    "initial.model",
						"endpoint": "initial.completion_endpoint",
					},
				},
			},
		})
	}

	return Pipeline{ID: id, Description: "retrieval augmented generation", Groups: groups}
}

// DefaultRAGPipeline constructs a pipeline using zero options for callers that
// do not need custom defaults.
func DefaultRAGPipeline(id string) Pipeline {
	return BuildRAGPipeline(id, RAGPipelineOptions{})
}

type ContextDocument struct {
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Score    float64                `json:"score,omitempty"`
}

type RAGResponse struct {
	Query            string            `json:"query"`
	Answer           string            `json:"answer"`
	Documents        []ContextDocument `json:"documents"`
	Model            string            `json:"model,omitempty"`
	Prompt           string            `json:"prompt,omitempty"`
	Context          string            `json:"context,omitempty"`
	ContextTruncated bool              `json:"context_truncated,omitempty"`
	Reasoning        string            `json:"reasoning,omitempty"`
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
	prompt, _ := data["build_prompt.default_output"].(map[string]interface{})
	ctxMap, _ := data["build_context.default_output"].(map[string]interface{})
	formattedCtx, _ := ctxMap["retrieved_context"].(string)
	truncated, _ := ctxMap["truncated"].(bool)
	query, _ := data["initial.query"].(string)
	model, _ := data["initial.model"].(string)
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
	reasonOut, _ := data["generate_reasoning.default_output"].(map[string]interface{})
	reasoning, _ := reasonOut["completion"].(string)
	pr := ""
	if prompt != nil {
		pr, _ = prompt["prompt"].(string)
	}
	return RAGResponse{Query: query, Answer: answer, Documents: docs, Model: model, Prompt: pr, Context: formattedCtx, ContextTruncated: truncated, Reasoning: reasoning}, true
}
