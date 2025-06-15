package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RemoteRerankProvider calls an external service to obtain rerank scores.
type RemoteRerankProvider struct {
	Endpoint string
	Client   *http.Client
	// MaxRetries controls how many attempts are made on failure.
	MaxRetries int
	Headers    map[string]string
}

// RemoteRerankOption customises a RemoteRerankProvider.
type RemoteRerankOption func(*RemoteRerankProvider)

// WithRerankHeader adds a header to outbound requests.
func WithRerankHeader(k, v string) RemoteRerankOption {
	return func(r *RemoteRerankProvider) {
		if r.Headers == nil {
			r.Headers = map[string]string{}
		}
		r.Headers[k] = v
	}
}

// NewRemoteRerankProvider creates a provider hitting the given endpoint.
func NewRemoteRerankProvider(endpoint string, opts ...RemoteRerankOption) *RemoteRerankProvider {
	p := &RemoteRerankProvider{
		Endpoint:   endpoint,
		Client:     &http.Client{Timeout: 30 * time.Second},
		MaxRetries: 2,
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

func (r *RemoteRerankProvider) Rerank(ctx context.Context, query string, docs []string) ([]float64, error) {
	payload := map[string]interface{}{"query": query, "documents": docs}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.Endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		for k, v := range r.Headers {
			req.Header.Set(k, v)
		}
		resp, err := r.Client.Do(req)
		if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var out struct {
				Scores []float64 `json:"scores"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				return nil, err
			}
			return out.Scores, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		if ctx.Err() != nil || attempt == r.MaxRetries {
			if err != nil {
				return nil, err
			}
			if resp != nil {
				return nil, fmt.Errorf("rerank service returned %s", resp.Status)
			}
			return nil, fmt.Errorf("rerank request failed")
		}
		backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
		time.Sleep(backoff)
	}
	return nil, fmt.Errorf("unreachable")
}
