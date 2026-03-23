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

	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
