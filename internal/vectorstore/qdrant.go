package vectorstore

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// QdrantStore implements VectorStore backed by a Qdrant server via HTTP API.
type QdrantStore struct {
	Endpoint   string
	Collection string
	Client     *http.Client
	APIKey     string
	Insecure   bool
}

// QdrantOption customizes QdrantStore creation.
type QdrantOption func(*QdrantStore)

// WithAPIKey sets the API key for authenticating requests.
func WithAPIKey(key string) QdrantOption {
	return func(q *QdrantStore) { q.APIKey = key }
}

// WithInsecureSkipVerify disables TLS certificate verification.
func WithInsecureSkipVerify() QdrantOption {
	return func(q *QdrantStore) { q.Insecure = true }
}

// NewQdrantStore constructs a store for the given endpoint and collection.
// Options allow configuring authentication and TLS behaviour.
func NewQdrantStore(endpoint, collection string, opts ...QdrantOption) *QdrantStore {
	qs := &QdrantStore{
		Endpoint:   endpoint,
		Collection: collection,
	}
	for _, opt := range opts {
		opt(qs)
	}
	tr := &http.Transport{}
	if qs.Insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	qs.Client = &http.Client{Timeout: 30 * time.Second, Transport: tr}
	return qs
}

func (q *QdrantStore) Upsert(ctx context.Context, docs []Document) error {
	points := make([]map[string]interface{}, len(docs))
	for i, d := range docs {
		points[i] = map[string]interface{}{
			"id":      d.ID,
			"vector":  d.Embedding,
			"payload": d.Metadata,
		}
	}
	payload := map[string]interface{}{"points": points}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/collections/%s/points?wait=true", q.Endpoint, q.Collection)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if q.APIKey != "" {
		req.Header.Set("api-key", q.APIKey)
	}

	resp, err := q.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("qdrant upsert error: %s", resp.Status)
	}
	return nil
}

func (q *QdrantStore) Query(ctx context.Context, embedding []float64, k int) ([]Document, error) {
	reqBody := map[string]interface{}{
		"vector": embedding,
		"limit":  k,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/collections/%s/points/search", q.Endpoint, q.Collection)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if q.APIKey != "" {
		req.Header.Set("api-key", q.APIKey)
	}

	resp, err := q.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qdrant search error: %s", resp.Status)
	}

	var out struct {
		Result []struct {
			ID      string                 `json:"id"`
			Score   float64                `json:"score"`
			Payload map[string]interface{} `json:"payload"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	docs := make([]Document, len(out.Result))
	for i, r := range out.Result {
		docs[i] = Document{ID: fmt.Sprint(r.ID), Embedding: nil, Metadata: r.Payload, Score: r.Score}
	}
	return docs, nil
}
