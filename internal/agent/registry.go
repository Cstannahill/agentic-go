package agent

import "sync"

// factory function type
type factory func() Agent

var (
	registry = make(map[string]factory)
	regMu    sync.RWMutex
)

// Register makes an agent constructor available by name.
func Register(name string, fn factory) {
	regMu.Lock()
	defer regMu.Unlock()
	registry[name] = fn
}

// New creates an agent by registered name.
func New(name string) (Agent, bool) {
	regMu.RLock()
	fn, ok := registry[name]
	regMu.RUnlock()
	if !ok {
		return nil, false
	}
	return fn(), true
}
