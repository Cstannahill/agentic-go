package vectorstore

import (
	"bytes"
	"context"
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
}

// NewQdrantStore constructs a store for the given endpoint and collection.
func NewQdrantStore(endpoint, collection string) *QdrantStore {
	return &QdrantStore{
		Endpoint:   endpoint,
		Collection: collection,
		Client:     &http.Client{Timeout: 30 * time.Second},
	}
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
		docs[i] = Document{ID: fmt.Sprint(r.ID), Embedding: nil, Metadata: r.Payload}
	}
	return docs, nil
}
