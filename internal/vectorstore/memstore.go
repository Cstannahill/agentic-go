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

// Query returns the k most similar documents.
func (m *MemoryStore) Query(ctx context.Context, emb []float64, k int) ([]Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	type scored struct {
		doc   Document
		score float64
	}
	var scoredDocs []scored
	for _, d := range m.docs {
		s := cosineSimilarity(d.Embedding, emb)
		scoredDocs = append(scoredDocs, scored{doc: d, score: s})
	}
	sort.Slice(scoredDocs, func(i, j int) bool { return scoredDocs[i].score > scoredDocs[j].score })
	if k > len(scoredDocs) {
		k = len(scoredDocs)
	}
	result := make([]Document, 0, k)
	for i := 0; i < k; i++ {
		result = append(result, scoredDocs[i].doc)
	}
	return result, nil
}
