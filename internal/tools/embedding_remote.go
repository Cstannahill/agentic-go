package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RemoteEmbeddingProvider calls an external HTTP service to generate embeddings.
// The service is expected to accept a JSON payload {"text": "..."} and respond
// with {"embedding": [..]}.
type RemoteEmbeddingProvider struct {
	Endpoint string
	Client   *http.Client
	// MaxRetries defines how many times a request should be retried on
	// transport errors or non-2xx responses.
	MaxRetries int
	Headers    map[string]string
}

// RemoteEmbedOption customises a RemoteEmbeddingProvider.
type RemoteEmbedOption func(*RemoteEmbeddingProvider)

// WithEmbedHeader adds a custom HTTP header to requests.
func WithEmbedHeader(k, v string) RemoteEmbedOption {
	return func(r *RemoteEmbeddingProvider) {
		if r.Headers == nil {
			r.Headers = map[string]string{}
		}
		r.Headers[k] = v
	}
}

// NewRemoteEmbeddingProvider constructs a provider targeting the given endpoint.
func NewRemoteEmbeddingProvider(endpoint string, opts ...RemoteEmbedOption) *RemoteEmbeddingProvider {
	p := &RemoteEmbeddingProvider{
		Endpoint:   endpoint,
		Client:     &http.Client{Timeout: 30 * time.Second},
		MaxRetries: 2,
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

func (r *RemoteEmbeddingProvider) Embed(ctx context.Context, text string) ([]float64, error) {
	payload := map[string]string{"text": text}
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
				Embedding []float64 `json:"embedding"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				return nil, err
			}
			return out.Embedding, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		if ctx.Err() != nil || attempt == r.MaxRetries {
			if err != nil {
				return nil, err
			}
			if resp != nil {
				return nil, fmt.Errorf("embedding service returned %s", resp.Status)
			}
			return nil, fmt.Errorf("embedding request failed")
		}
		backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
		time.Sleep(backoff)
	}
	return nil, fmt.Errorf("unreachable")
}
