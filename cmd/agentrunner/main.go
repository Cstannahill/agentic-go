package main

import (
	"context"
	"fmt"
	"time"

	"agentic.example.com/mvp/internal/agent"
	"agentic.example.com/mvp/internal/orchestrator"
)

func main() {
	fmt.Println("--- Agentic Flow Engine Runner ---")

	orc := orchestrator.NewOrchestrator()

	pipeline := orchestrator.Pipeline{
		ID:          "simple_echo_pipeline_001",
		Description: "A pipeline with two echo steps",
		Groups: []orchestrator.PipelineGroup{
			{
				Name: "initial",
				Steps: []orchestrator.PipelineStep{
					{
						Name:      "step_one_echo",
						AgentType: "EchoAgent",
						AgentConfig: agent.Task{
							Description: "First echo in the pipeline",
						},
						InputMappings: map[string]string{
							"message": "initial.user_greeting",
							"detail":  "initial.user_detail",
						},
					},
					{
						Name:      "step_two_echo",
						AgentType: "EchoAgent",
						AgentConfig: agent.Task{
							Description: "Second echo, uses output from step one",
						},
						InputMappings: map[string]string{
							"complex_input":     fmt.Sprintf("step_one_echo.%s", orchestrator.DefaultOutputKey),
							"original_greeting": "initial.user_greeting",
						},
					},
				},
			},
		},
	}

	initialInput := map[string]interface{}{
		"user_greeting": "Hello from the pipeline!",
		"user_detail":   "This is extra detail for step one.",
	}

	fmt.Printf("\nExecuting pipeline ID: %s\n", pipeline.ID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	finalData, err := orc.ExecutePipeline(ctx, pipeline, initialInput)
	if err != nil {
		fmt.Printf("Pipeline execution failed: %v\n", err)
	} else {
		fmt.Printf("Pipeline executed successfully!\nFinal State: %v\n", finalData)
	}

	fmt.Println("--- Agentic Flow Engine Runner Finished ---")
}
