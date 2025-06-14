package main

import (
	"encoding/json"
	"log"
	"net/http"

	"agentic.example.com/mvp/internal/config"
	"agentic.example.com/mvp/internal/orchestrator"
	"agentic.example.com/mvp/internal/tools"
	"agentic.example.com/mvp/internal/vectorstore"
)

// executeRequest is the payload for POST /execute
type executeRequest struct {
	Pipeline     orchestrator.Pipeline  `json:"pipeline"`
	InitialInput map[string]interface{} `json:"initial_input"`
}

type executeResponse struct {
	Data  orchestrator.StepData `json:"data,omitempty"`
	Error string                `json:"error,omitempty"`
}

func main() {
	cfg := config.LoadFromEnv()

	vectorstore.InitDefault(cfg.VectorStore)
	tools.InitDefaults(cfg)

	orc := orchestrator.NewOrchestrator()

	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req executeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data, err := orc.ExecutePipeline(r.Context(), req.Pipeline, req.InitialInput)
		resp := executeResponse{Data: data}
		if err != nil {
			resp.Error = err.Error()
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
