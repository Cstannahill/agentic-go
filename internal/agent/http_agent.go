package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// HTTPCallAgent is an agent that makes HTTP requests.
// It demonstrates how external services or tools can be plugged into
// the orchestration layer in a code-agnostic way.
type HTTPCallAgent struct {
	agentID string
	client  *http.Client
	Method  string
	URL     string
	Headers map[string]string
}

// NewHTTPCallAgent creates a new HTTPCallAgent with the provided configuration.
func NewHTTPCallAgent(method, url string, headers map[string]string) *HTTPCallAgent {
	return &HTTPCallAgent{
		agentID: fmt.Sprintf("http-agent-%s", uuid.NewString()),
		client:  &http.Client{Timeout: 30 * time.Second},
		Method:  method,
		URL:     url,
		Headers: headers,
	}
}

func (ha *HTTPCallAgent) ID() string { return ha.agentID }

// Execute sends the task input as JSON to the configured endpoint and returns the response body.
func (ha *HTTPCallAgent) Execute(ctx context.Context, task Task) Result {
	body, err := json.Marshal(task.Input)
	if err != nil {
		return Result{TaskID: task.ID, Error: err, Successful: false}
	}

	req, err := http.NewRequestWithContext(ctx, ha.Method, ha.URL, bytes.NewReader(body))
	if err != nil {
		return Result{TaskID: task.ID, Error: err, Successful: false}
	}
	for k, v := range ha.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ha.client.Do(req)
	if err != nil {
		return Result{TaskID: task.ID, Error: err, Successful: false}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{TaskID: task.ID, Error: err, Successful: false}
	}

	return Result{TaskID: task.ID, Output: string(respBody), Successful: true}
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

// HTTPCallAgent calls an HTTP endpoint with provided parameters.
type HTTPCallAgent struct {
	agentID string
}

func NewHTTPCallAgent() *HTTPCallAgent {
	return &HTTPCallAgent{agentID: fmt.Sprintf("http-agent-%s", uuid.NewString())}
}

func (h *HTTPCallAgent) ID() string { return h.agentID }

func (h *HTTPCallAgent) Execute(ctx context.Context, task Task) Result {
	url, _ := task.Input["url"].(string)
	method, _ := task.Input["method"].(string)
	bodyStr, _ := task.Input["body"].(string)
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBufferString(bodyStr))
	if err != nil {
		return Result{TaskID: task.ID, Error: err}
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Result{TaskID: task.ID, Error: err}
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{TaskID: task.ID, Error: err}
	}
	return Result{TaskID: task.ID, Output: map[string]interface{}{"status": resp.StatusCode, "body": string(data)}, Successful: true}
}

func init() {
	Register("HTTPCallAgent", func() Agent { return NewHTTPCallAgent() })
}
