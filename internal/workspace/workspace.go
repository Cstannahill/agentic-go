package workspace

import "sync"

// Store provides a simple in-memory shared workspace for agents.
type Store struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// NewStore creates an empty Store.
func NewStore() *Store {
	return &Store{data: make(map[string][]byte)}
}

// DefaultStore is used when callers do not provide their own Store.
var DefaultStore = NewStore()

// Put stores a value under the given key.
func (s *Store) Put(key string, val []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
}

// Get retrieves a value by key.
func (s *Store) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// Keys returns a copy of all keys currently in the store.
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}
