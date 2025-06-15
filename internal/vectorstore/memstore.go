package vectorstore

import (
	"context"
	"math"
	"sort"
	"sync"
)

// MemoryStore provides an in-memory implementation of VectorStore.
type MemoryStore struct {
	mu   sync.RWMutex
	docs []Document
}

// NewMemoryStore returns a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// Upsert adds or replaces documents based on ID.
func (m *MemoryStore) Upsert(ctx context.Context, docs []Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range docs {
		replaced := false
		for i, existing := range m.docs {
			if existing.ID == d.ID {
				m.docs[i] = d
				replaced = true
				break
			}
		}
		if !replaced {
			m.docs = append(m.docs, d)
		}
	}
	return nil
}

// cosineSimilarity returns similarity between two vectors.
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func matchesFilter(meta map[string]interface{}, filter map[string]interface{}) bool {
	if len(filter) == 0 {
		return true
	}
	for k, v := range filter {
		if meta[k] != v {
			return false
		}
	}
	return true
}

// Query returns the most similar documents subject to an optional filter.
func (m *MemoryStore) Query(ctx context.Context, req QueryRequest) ([]Document, error) {
	emb := req.Embedding
	k := req.TopK
	m.mu.RLock()
	defer m.mu.RUnlock()
	type scored struct {
		doc   Document
		score float64
	}
	var scoredDocs []scored
	for _, d := range m.docs {
		if !matchesFilter(d.Metadata, req.Filter) {
			continue
		}
		s := cosineSimilarity(d.Embedding, emb)
		scoredDocs = append(scoredDocs, scored{doc: d, score: s})
	}
	sort.Slice(scoredDocs, func(i, j int) bool { return scoredDocs[i].score > scoredDocs[j].score })
	if k > len(scoredDocs) {
		k = len(scoredDocs)
	}
	result := make([]Document, 0, k)
	for i := 0; i < k; i++ {
		d := scoredDocs[i].doc
		d.Score = scoredDocs[i].score
		result = append(result, d)
	}
	return result, nil
}

// Delete removes documents with the specified IDs.
func (m *MemoryStore) Delete(ctx context.Context, ids []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	filtered := m.docs[:0]
	for _, d := range m.docs {
		if _, ok := idSet[d.ID]; !ok {
			filtered = append(filtered, d)
		}
	}
	m.docs = filtered
	return nil
}
