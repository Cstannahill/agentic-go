package orchestrator

import (
	"context"
	"fmt"

	"agentic.example.com/mvp/internal/watcher"
)

// StartWatcher begins consuming events from the provided watcher and launches
// a pipeline run for each event. A bounded work queue is used to avoid
// unbounded goroutines when events arrive faster than pipelines can start.
func (o *Orchestrator) StartWatcher(ctx context.Context, w watcher.Watcher, p Pipeline, queueSize int) error {
	if queueSize <= 0 {
		queueSize = 1
	}
	if err := w.Start(ctx); err != nil {
		return err
	}

	work := make(chan map[string]interface{}, queueSize)

	go func() {
		defer close(work)
		for {
			select {
			case ev, ok := <-w.Events():
				if !ok {
					return
				}
				select {
				case work <- ev.Payload:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case payload, ok := <-work:
				if !ok {
					return
				}
				go func(data map[string]interface{}) {
					if _, err := o.ExecutePipeline(context.Background(), p, data); err != nil {
						fmt.Printf("Watcher pipeline error: %v\n", err)
					}
				}(payload)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}
