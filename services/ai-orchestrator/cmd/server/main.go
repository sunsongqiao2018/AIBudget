package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/adapter/llm"
	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/app/categorize"
	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/app/chat"
	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/domain"
	httphandler "github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/transport/http"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Determine which LLM provider to use based on environment configuration.
	// Set LLM_PROVIDER=anthropic to use the real Anthropic Claude API.
	// Otherwise, the mock provider is used (suitable for development/testing).
	llmProviderName := os.Getenv("LLM_PROVIDER")
	llmModel := os.Getenv("LLM_MODEL")
	anthropicAPIKey := os.Getenv("ANTHROPIC_API_KEY")

	var llmProvider domain.LLMProvider

	switch llmProviderName {
	case "anthropic":
		timeoutSecs := 30
		if v := os.Getenv("LLM_TIMEOUT_SECS"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				timeoutSecs = n
			}
		}
		cfg := domain.ProviderConfig{
			Provider:    "anthropic",
			APIKey:      anthropicAPIKey,
			Model:       llmModel,
			TimeoutSecs: timeoutSecs,
		}
		llmProvider = llm.NewAnthropicProvider(cfg)
		log.Printf("Using Anthropic LLM provider (model: %s)", cfg.Model)

	default:
		llmProvider = llm.NewMockProvider()
		log.Printf("Using mock LLM provider (ready to process requests)")
	}

	// Create application services
	categorizeService := categorize.NewService(llmProvider)
	chatService := chat.NewService(llmProvider)

	// Create HTTP handler with dependencies
	handler := httphandler.NewHandler(categorizeService, chatService)

	// Register routes
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	log.Printf("ai-orchestrator listening on :%s", port)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
