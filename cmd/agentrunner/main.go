// cmd/agentrunner/main.go
package main

import (
	"context"
	"fmt"
	"sync" // Import the sync package for WaitGroup
	"time"

	"agentic.example.com/mvp/internal/agent"
)

func main() {
	fmt.Println("--- Agent Runner Starting (Batch Processing Edition) ---")

	echoAgent := agent.NewEchoAgent()
	var genericAgent agent.Agent = echoAgent

	fmt.Printf("Created Agent. ID: %s\n", genericAgent.ID())

	// Define a list of tasks
	tasks := []agent.Task{
		{
			ID:          "batch-task-001",
			Description: "First task in batch",
			Input:       map[string]interface{}{"message": "Hello Batch 1", "delay_ms": 800},
		},
		{
			ID:          "batch-task-002",
			Description: "Second task in batch (will be quick)",
			Input:       map[string]interface{}{"message": "Hello Batch 2", "delay_ms": 300},
		},
		{
			ID:          "batch-task-003",
			Description: "Third task in batch (will timeout)",
			Input:       map[string]interface{}{"message": "Hello Batch 3 - timeout", "delay_ms": 2000},
		},
		{
			ID:          "batch-task-004",
			Description: "Fourth task in batch",
			Input:       map[string]interface{}{"message": "Hello Batch 4", "delay_ms": 500},
		},
	}

	numTasks := len(tasks)
	// Create a channel to receive results.
	// For simplicity, we'll make it unbuffered. We could also make it buffered:
	// resultChannel := make(chan agent.Result, numTasks) // Buffered channel
	resultChannel := make(chan agent.Result)

	// --- Introducing sync.WaitGroup ---
	var wg sync.WaitGroup

	// Context for tasks that might timeout
	// Task 3 is designed to take 2000ms, this context will timeout at 1 second.
	taskCtx, taskCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer taskCancel() // Ensure all resources for this context are cleaned up

	fmt.Printf("Dispatching %d tasks...\n", numTasks)

	for i, task := range tasks {
		// Increment the WaitGroup counter *before* launching the goroutine.
		wg.Add(1)

		// Important: When launching goroutines in a loop, pass loop variables
		// as arguments to the goroutine's function to avoid capturing the wrong value.
		// Each goroutine needs its own copy of 'task' and 'i' (if 'i' were used inside).
		go func(currentTask agent.Task, taskIndex int) {
			// Decrement the counter when the goroutine finishes, using defer.
			defer wg.Done()

			fmt.Printf("[%s] Goroutine for task '%s' (index %d) started.\n", genericAgent.ID(), currentTask.ID, taskIndex)

			var individualTaskCtx context.Context
			var individualCancel context.CancelFunc // To store cancel func if needed

			// Task 3 (index 2) is specifically designed to use the shorter timeout context
			if currentTask.ID == "batch-task-003" {
				individualTaskCtx = taskCtx // Use the context that times out sooner
			} else {
				// For other tasks, give them a longer, more "general" timeout for their execution
				// Or use context.Background() if no specific shorter timeout is needed for them beyond a global one.
				// For this example, let's give them a slightly longer default timeout.
				individualTaskCtx, individualCancel = context.WithTimeout(context.Background(), 3*time.Second)
				if individualCancel != nil { // Important to cancel these individual contexts too if created
					defer individualCancel()
				}
			}
			
			result := genericAgent.Execute(individualTaskCtx, currentTask)
			
			fmt.Printf("[%s] Goroutine for task '%s' sending result...\n", genericAgent.ID(), currentTask.ID)
			resultChannel <- result
			fmt.Printf("[%s] Goroutine for task '%s' finished sending result.\n", genericAgent.ID(), currentTask.ID)

		}(task, i) // Pass task and i to the goroutine
	}

	// Goroutine to wait for all tasks to complete and then close the channel
	// This is a common pattern to signal the receiver that no more results will be sent.
	go func() {
		fmt.Println("Coordinator goroutine: Waiting for all worker goroutines to complete...")
		wg.Wait() // Wait for all goroutines (wg.Done() calls)
		close(resultChannel) // Close the channel once all workers are done
		fmt.Println("Coordinator goroutine: All workers done. Result channel closed.")
	}()

	fmt.Println("Main goroutine: Collecting results...")
	// Collect results. Since we closed the channel, we can use a for...range loop.
	// The loop will automatically break when the channel is closed.
	var receivedResults []agent.Result
	for result := range resultChannel {
		fmt.Printf("Main goroutine: Received result for Task ID: %s (Success: %t)\n", result.TaskID, result.Successful)
		if !result.Successful {
			fmt.Printf("  Error for Task ID %s: %v\n", result.TaskID, result.Error)
		}
		receivedResults = append(receivedResults, result)
	}

	fmt.Printf("\n--- All %d tasks processed. Summary: ---\n", len(receivedResults))
	for _, res := range receivedResults {
		status := "SUCCESS"
		errMsg := ""
		if !res.Successful {
			status = "FAILED"
			errMsg = fmt.Sprintf(" (Error: %v)", res.Error)
		}
		fmt.Printf("Task ID: %s, Status: %s%s\n", res.TaskID, status, errMsg)
		// fmt.Printf("  Output: %v\n", res.Output) // Optionally print full output
	}

	fmt.Println("--- Agent Runner Finished ---")
}