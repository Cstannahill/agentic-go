// internal/agent/agent.go
package agent

import (
	"context"
	"fmt"  // For formatting strings
	"time" // For simulating work

	"github.com/google/uuid" // For generating unique IDs
)

// Task represents a piece of work for an agent to perform.
// For now, it's simple, but it can be expanded.
type Task struct {
	ID          string
	Description string
	Input       map[string]interface{} // Flexible input for the task
}

// Result holds the output of a task performed by an agent.
type Result struct {
	TaskID     string
	Output     interface{}
	Error      error // If the task failed, the error will be here
	Successful bool
}

// Agent defines the contract for any autonomous worker in our system.
// An agent is responsible for executing tasks.
type Agent interface {
	// ID returns a unique identifier for the agent.
	ID() string

	// Execute performs a given task.
	// It takes a context (for managing cancellation, timeouts, etc. - crucial for agentic systems)
	// and a Task as input.
	// It returns a Result containing the task's output or an error.
	Execute(ctx context.Context, task Task) Result

	// TODO: Add more methods as needed, e.g.:
	// Status() string // To get the current status of the agent (idle, busy, error)
	// Stop() error    // To gracefully stop an agent
	// Initialize(config map[string]interface{}) error // To set up an agent
}


// EchoAgent is a simple agent that echoes back the input it receives.
type EchoAgent struct {
	agentID string
	// TODO: Add other agent-specific state here, e.g., configuration, logger
}

// NewEchoAgent creates and returns a new EchoAgent.
// This is a common Go pattern called a "constructor function".
// Go doesn't have classes or constructors in the C#/Java sense.
// Instead, you typically export a function (e.g., New<TypeName>) to create instances of your struct.
func NewEchoAgent() *EchoAgent { // It returns a pointer to an EchoAgent
	return &EchoAgent{
		agentID: fmt.Sprintf("echo-agent-%s", uuid.NewString()),
	}
}

// ID implements the Agent interface.
// This is a "method" on the EchoAgent struct.
// Note the receiver: `(ea *EchoAgent)`. This means the method operates on an instance of EchoAgent.
// Using a pointer receiver `*EchoAgent` is common if the method needs to modify the struct's state
// or if the struct is large and you want to avoid copying it.
func (ea *EchoAgent) ID() string {
	return ea.agentID
}

// Execute implements the Agent interface for EchoAgent.
func (ea *EchoAgent) Execute(ctx context.Context, task Task) Result {
	fmt.Printf("[%s] Received task: %s - Input: %v\n", ea.agentID, task.Description, task.Input)

	// Get desired delay from task input, default to 1 second
	simulatedWorkDuration := 1 * time.Second // Default duration
	if delay, ok := task.Input["delay_ms"].(int); ok { // Type assertion to int
		simulatedWorkDuration = time.Duration(delay) * time.Millisecond
	} else if delay, ok := task.Input["delay_ms"].(float64); ok { // Handle if number is float64 (common from JSON)
		simulatedWorkDuration = time.Duration(delay) * time.Millisecond
	}


	fmt.Printf("[%s] Simulating work for %v...\n", ea.agentID, simulatedWorkDuration)
	select {
	case <-time.After(simulatedWorkDuration): // Use the determined duration
		// Work done
		fmt.Printf("[%s] Simulated work finished for task: %s.\n", ea.agentID, task.Description)
	case <-ctx.Done():
		fmt.Printf("[%s] Task execution cancelled: %s (Reason: %v)\n", ea.agentID, task.Description, ctx.Err())
		return Result{
			TaskID:     task.ID,
			Output:     nil,
			Error:      ctx.Err(),
			Successful: false,
		}
	}

	outputMessage := fmt.Sprintf("Echoing from %s: Processed task '%s'", ea.agentID, task.Description)
	processedInput := make(map[string]interface{})
	for k, v := range task.Input {
		processedInput[k] = fmt.Sprintf("Echoed: %v", v)
	}

	fmt.Printf("[%s] Finished task processing: %s\n", ea.agentID, task.Description)
	return Result{
		TaskID:     task.ID,
		Output:     map[string]interface{}{"message": outputMessage, "processed_input": processedInput},
		Error:      nil,
		Successful: true,
	}
}

// --- Side Note on `context.Context` ---
// `context.Context` is a standard Go package used to carry deadlines, cancellation signals,
// and other request-scoped values across API boundaries and between goroutines.
// It's idiomatic in Go to pass a `Context` as the first argument to functions that
// might involve I/O, long-running computations, or calls to external services.
// For agentic systems, where tasks might be long-running or need to be cancelled,
// using `context` from the start is a very good practice.
// Think of it a bit like CancellationToken in C# or AbortSignal in web APIs.