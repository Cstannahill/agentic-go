package agent

import (
	"bytes"
	"context"
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
