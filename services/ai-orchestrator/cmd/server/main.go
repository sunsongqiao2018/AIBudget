package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/adapter/llm"
	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/app/categorize"
	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/app/chat"
	httphandler "github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/transport/http"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Create LLM provider (using mock for MVP)
	// In production, this would be configurable (OpenAI, Anthropic, etc.)
	llmProvider := llm.NewMockProvider()

	// Create application services
	categorizeService := categorize.NewService(llmProvider)
	chatService := chat.NewService(llmProvider)

	// Create HTTP handler with dependencies
	handler := httphandler.NewHandler(categorizeService, chatService)

	// Register routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	log.Printf("ai-orchestrator listening on :%s", port)
	log.Printf("Using mock LLM provider (ready to process requests)")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
