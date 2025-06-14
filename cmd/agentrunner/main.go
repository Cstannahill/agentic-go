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

	srv := &http.Server{Addr: ":8081"}
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})
	go srv.ListenAndServe()
	defer srv.Shutdown(context.Background())

	orc := orchestrator.NewOrchestrator()

	pipeline := orchestrator.Pipeline{
		ID:          "simple_echo_pipeline_001",
		Description: "A pipeline with echo and HTTP steps",
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

	events, errCh := orc.RunPipeline(ctx, pipeline, initialInput)
	finalData := make(orchestrator.StepData)
	for ev := range events {
		fmt.Printf("Event: step %s finished with output %v\n", ev.Step, ev.Result.Output)
		if ev.Result.Output != nil {
			finalData[fmt.Sprintf("%s.%s", ev.Step, orchestrator.DefaultOutputKey)] = ev.Result.Output
		}
		finalData[fmt.Sprintf("%s.successful", ev.Step)] = ev.Result.Successful
	}

	if err := <-errCh; err != nil {
		fmt.Printf("Pipeline execution failed: %v\n", err)
	} else {
		fmt.Printf("Pipeline executed successfully!\nFinal State: %v\n", finalData)
	}

	fmt.Println("--- Agentic Flow Engine Runner Finished ---")
}
