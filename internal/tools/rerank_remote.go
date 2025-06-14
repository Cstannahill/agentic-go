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
}

// NewRemoteRerankProvider creates a provider hitting the given endpoint.
func NewRemoteRerankProvider(endpoint string) *RemoteRerankProvider {
	return &RemoteRerankProvider{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (r *RemoteRerankProvider) Rerank(ctx context.Context, query string, docs []string) ([]float64, error) {
	payload := map[string]interface{}{"query": query, "documents": docs}
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
		return nil, fmt.Errorf("rerank service returned %s", resp.Status)
	}
	var out struct {
		Scores []float64 `json:"scores"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Scores, nil
}
