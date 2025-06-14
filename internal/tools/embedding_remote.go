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
}

// NewRemoteEmbeddingProvider constructs a provider targeting the given endpoint.
func NewRemoteEmbeddingProvider(endpoint string) *RemoteEmbeddingProvider {
	return &RemoteEmbeddingProvider{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (r *RemoteEmbeddingProvider) Embed(ctx context.Context, text string) ([]float64, error) {
	payload := map[string]string{"text": text}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned %s", resp.Status)
	}
	var out struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Embedding, nil
}
