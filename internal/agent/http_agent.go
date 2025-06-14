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

// HTTPCallAgent is an agent that makes HTTP requests to an external service.
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
}

func init() {
	Register("HTTPCallAgent", func() Agent { return NewHTTPCallAgent(http.MethodGet, "", nil) })
}
