// Package http provides HTTP handlers for the ai-orchestrator service.
package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/app/categorize"
	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/app/chat"
	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/domain"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	categorizeService *categorize.Service
	chatService       *chat.Service
}

// NewHandler creates a new HTTP handler with the given dependencies.
func NewHandler(categorizeService *categorize.Service, chatService *chat.Service) *Handler {
	return &Handler{
		categorizeService: categorizeService,
		chatService:       chatService,
	}
}

// RegisterRoutes registers all HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/internal/categorize", h.Categorize)
	mux.HandleFunc("/internal/chat/query", h.ChatQuery)
}

// Health returns the service health status.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "ai-orchestrator",
	})
}

// Categorize handles transaction categorization requests.
func (h *Handler) Categorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		respondError(w, http.StatusBadRequest, "content-type must be application/json")
		return
	}

	var req domain.CategorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	resp, err := h.categorizeService.Categorize(r.Context(), req)
	if err != nil {
		var llmErr *domain.LLMError
		if errors.As(err, &llmErr) {
			// Determine if it's a client error or server error
			if llmErr.Message == "no transactions provided" ||
				llmErr.Message == "transaction[" {
				respondError(w, http.StatusBadRequest, "validation error: "+llmErr.Message)
				return
			}
			respondError(w, http.StatusInternalServerError, "AI processing error: "+llmErr.Message)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// ChatQuery handles chat query requests.
func (h *Handler) ChatQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		respondError(w, http.StatusBadRequest, "content-type must be application/json")
		return
	}

	var req domain.ChatQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	resp, err := h.chatService.Query(r.Context(), req)
	if err != nil {
		var llmErr *domain.LLMError
		if errors.As(err, &llmErr) {
			// Determine if it's a client error or server error
			if llmErr.Message == "question is required" ||
				llmErr.Message == "question must be at least 3 characters" ||
				llmErr.Message == "ledgerSummary.rangeStart is required" ||
				llmErr.Message == "ledgerSummary.rangeEnd is required" {
				respondError(w, http.StatusBadRequest, "validation error: "+llmErr.Message)
				return
			}
			respondError(w, http.StatusInternalServerError, "AI processing error: "+llmErr.Message)
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// respondError writes a JSON error response.
func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(domain.ErrorResponse{Error: message})
}
