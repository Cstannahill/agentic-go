package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// CompletionTool sends a prompt to a remote LLM endpoint and returns the completion text.
// It is designed to work with the Universal MCP gateway.
type CompletionTool struct {
	Endpoint string
	Client   *http.Client
}

var defaultCompletionEndpoint = "http://localhost:8080/completion"

// SetDefaultCompletionEndpoint defines the endpoint used when none is provided.
func SetDefaultCompletionEndpoint(ep string) { defaultCompletionEndpoint = ep }

// DefaultCompletionEndpoint returns the currently configured default endpoint.
func DefaultCompletionEndpoint() string { return defaultCompletionEndpoint }

// NewCompletionTool creates a CompletionTool for the given endpoint.
func NewCompletionTool(endpoint string) *CompletionTool {
	return &CompletionTool{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: 60 * time.Second},
	}
}

// NewDefaultCompletionTool returns a tool using the configured default endpoint.
func NewDefaultCompletionTool() *CompletionTool {
	return NewCompletionTool(DefaultCompletionEndpoint())
}

func (c *CompletionTool) Run(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	prompt, ok := input["prompt"].(string)
	if !ok || prompt == "" {
		return nil, errors.New("prompt required")
	}
	model, _ := input["model"].(string)
	reqBody := map[string]interface{}{
		"prompt": prompt,
	}
	if model != "" {
		reqBody["model"] = model
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	var out struct {
		Completion string `json:"completion"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return map[string]interface{}{"completion": out.Completion}, nil
}
