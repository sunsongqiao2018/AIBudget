// Package http provides HTTP handlers for the edge-api service.
package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sunsongqiao2018/AIBudget/services/edge-api/internal/domain"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	aiClient domain.AIOrchestratorClient
}

// NewHandler creates a new HTTP handler with the given dependencies.
func NewHandler(aiClient domain.AIOrchestratorClient) *Handler {
	return &Handler{
		aiClient: aiClient,
	}
}

// RegisterRoutes registers all HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/v1/ai/categorize", h.Categorize)
	mux.HandleFunc("/v1/ai/chat/query", h.ChatQuery)
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
		"service": "edge-api",
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

	if len(req.Transactions) == 0 {
		respondError(w, http.StatusBadRequest, "transactions array must not be empty")
		return
	}

	// Validate each transaction
	for i, tx := range req.Transactions {
		if tx.Date == "" {
			respondError(w, http.StatusBadRequest, "transaction["+string(rune('0'+i))+"].date is required")
			return
		}
		if tx.Currency == "" {
			respondError(w, http.StatusBadRequest, "transaction["+string(rune('0'+i))+"].currency is required")
			return
		}
		if tx.Merchant == "" {
			respondError(w, http.StatusBadRequest, "transaction["+string(rune('0'+i))+"].merchant is required")
			return
		}
		if tx.Amount <= 0 {
			respondError(w, http.StatusBadRequest, "transaction["+string(rune('0'+i))+"].amount must be positive")
			return
		}
	}

	resp, err := h.aiClient.Categorize(r.Context(), req)
	if err != nil {
		var categorizeErr domain.CategorizeError
		if errors.As(err, &categorizeErr) {
			respondError(w, http.StatusBadGateway, "AI service error: "+categorizeErr.Message)
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

	if req.Question == "" {
		respondError(w, http.StatusBadRequest, "question is required")
		return
	}

	if req.LedgerSummary.RangeStart == "" || req.LedgerSummary.RangeEnd == "" {
		respondError(w, http.StatusBadRequest, "ledgerSummary.rangeStart and rangeEnd are required")
		return
	}

	resp, err := h.aiClient.ChatQuery(r.Context(), req)
	if err != nil {
		var chatErr domain.ChatQueryError
		if errors.As(err, &chatErr) {
			respondError(w, http.StatusBadGateway, "AI service error: "+chatErr.Message)
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
