package watcher

import (
	"context"
	"time"
)

// Event represents an external trigger that starts a pipeline run.
type Event struct {
	Payload map[string]interface{}
}

// Watcher emits events that can be consumed by the orchestrator.
type Watcher interface {
	// Events returns a channel of incoming events.
	Events() <-chan Event
	// Start begins producing events until the context is cancelled.
	Start(ctx context.Context) error
}

// TickerWatcher generates events on a fixed interval. It is useful for
// demonstration and testing when no real event source is available.
type TickerWatcher struct {
	Interval time.Duration
	events   chan Event
}

// NewTickerWatcher creates a watcher that emits an empty payload on each tick.
func NewTickerWatcher(interval time.Duration) *TickerWatcher {
	return &TickerWatcher{Interval: interval, events: make(chan Event)}
}

func (t *TickerWatcher) Events() <-chan Event { return t.events }

func (t *TickerWatcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(t.Interval)
	go func() {
		defer close(t.events)
		for {
			select {
			case tm := <-ticker.C:
				t.events <- Event{Payload: map[string]interface{}{"time": tm}}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	return nil
}
