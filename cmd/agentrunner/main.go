package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"agentic.example.com/mvp/internal/agent"
	"agentic.example.com/mvp/internal/orchestrator"
)

func main() {
	fmt.Println("--- Agentic Flow Engine Runner ---")

	// Start a lightweight local HTTP server for demo purposes.
	srv := &http.Server{Addr: ":8081"}
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})
	go srv.ListenAndServe()
	defer srv.Shutdown(context.Background())

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
				Name:      "step_two_http",
				AgentType: "HTTPCallAgent",
				AgentConfig: agent.Task{
					Description: "Call local HTTP service",
				},
				InputMappings: map[string]string{
					"url":    "initial.ping_url",
					"method": "initial.http_method",

				},
			},
		},
	}

	initialInput := map[string]interface{}{
		"user_greeting": "Hello from the pipeline!",
		"user_detail":   "This is extra detail for step one.",
		"ping_url":      "http://localhost:8081/ping",
		"http_method":   "GET",
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
