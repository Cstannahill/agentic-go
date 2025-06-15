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
}

// NewRemoteEmbeddingProvider constructs a provider targeting the given endpoint.
func NewRemoteEmbeddingProvider(endpoint string) *RemoteEmbeddingProvider {
	return &RemoteEmbeddingProvider{
		Endpoint:   endpoint,
		Client:     &http.Client{Timeout: 30 * time.Second},
		MaxRetries: 2,
	}
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
		time.Sleep(time.Duration(attempt+1) * 100 * time.Millisecond)
	}
	return nil, fmt.Errorf("unreachable")
}
