package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sunsongqiao2018/AIBudget/services/edge-api/internal/adapter/ai"
	httphandler "github.com/sunsongqiao2018/AIBudget/services/edge-api/internal/transport/http"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	aiOrchestratorURL := os.Getenv("AI_ORCHESTRATOR_URL")
	if aiOrchestratorURL == "" {
		aiOrchestratorURL = "http://localhost:8081"
	}

	// Create AI orchestrator client
	aiClient := ai.NewClient(aiOrchestratorURL, 30*time.Second)

	// Create HTTP handler with dependencies
	handler := httphandler.NewHandler(aiClient)

	// Register routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	log.Printf("edge-api listening on :%s", port)
	log.Printf("AI orchestrator URL: %s", aiOrchestratorURL)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
