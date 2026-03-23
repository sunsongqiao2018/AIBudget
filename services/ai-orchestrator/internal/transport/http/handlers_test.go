package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/app/categorize"
	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/app/chat"
	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/domain"
)

// mockLLMProvider is a test implementation of domain.LLMProvider.
type mockLLMProvider struct {
	categorizeFunc func(ctx context.Context, candidates []domain.TransactionCandidate) ([]domain.CategorizedTransaction, error)
	chatQueryFunc  func(ctx context.Context, question string, ledger domain.LedgerSummary) (domain.ChatResponse, error)
}

func (m *mockLLMProvider) Categorize(ctx context.Context, candidates []domain.TransactionCandidate) ([]domain.CategorizedTransaction, error) {
	if m.categorizeFunc != nil {
		return m.categorizeFunc(ctx, candidates)
	}
	return nil, nil
}

func (m *mockLLMProvider) ChatQuery(ctx context.Context, question string, ledger domain.LedgerSummary) (domain.ChatResponse, error) {
	if m.chatQueryFunc != nil {
		return m.chatQueryFunc(ctx, question, ledger)
	}
	return domain.ChatResponse{}, nil
}

func setupTestHandler() (*Handler, *mockLLMProvider) {
	mockProvider := &mockLLMProvider{
		categorizeFunc: func(ctx context.Context, candidates []domain.TransactionCandidate) ([]domain.CategorizedTransaction, error) {
			results := make([]domain.CategorizedTransaction, len(candidates))
			for i, c := range candidates {
				results[i] = domain.CategorizedTransaction{
					TransactionCandidate: c,
					Category:             "Test Category",
					Confidence:           0.9,
				}
			}
			return results, nil
		},
		chatQueryFunc: func(ctx context.Context, question string, ledger domain.LedgerSummary) (domain.ChatResponse, error) {
			return domain.ChatResponse{
				Answer: "Test answer",
				Citations: []domain.MetricCitation{
					{Label: "Test", Value: "Value"},
				},
			}, nil
		},
	}

	categorizeService := categorize.NewService(mockProvider)
	chatService := chat.NewService(mockProvider)

	return NewHandler(categorizeService, chatService), mockProvider
}

func TestHandler_Health(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("status: got %q, want %q", resp["status"], "ok")
	}

	if resp["service"] != "ai-orchestrator" {
		t.Errorf("service: got %q, want %q", resp["service"], "ai-orchestrator")
	}
}

func TestHandler_Health_MethodNotAllowed(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()

	handler.Health(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandler_Categorize_Success(t *testing.T) {
	handler, _ := setupTestHandler()

	reqBody := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Test"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/internal/categorize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Categorize(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusOK)
	}

	var resp domain.CategorizeResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	if resp.Results[0].Category != "Test Category" {
		t.Errorf("Category: got %q, want %q", resp.Results[0].Category, "Test Category")
	}
}

func TestHandler_Categorize_EmptyTransactions(t *testing.T) {
	handler, _ := setupTestHandler()

	reqBody := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/internal/categorize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Categorize(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandler_Categorize_InvalidJSON(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/internal/categorize", bytes.NewReader([]byte(`invalid`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Categorize(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandler_Categorize_MethodNotAllowed(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/internal/categorize", nil)
	rr := httptest.NewRecorder()

	handler.Categorize(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandler_Categorize_ProviderError(t *testing.T) {
	mockProvider := &mockLLMProvider{
		categorizeFunc: func(ctx context.Context, candidates []domain.TransactionCandidate) ([]domain.CategorizedTransaction, error) {
			return nil, errors.New("LLM service unavailable")
		},
	}

	categorizeService := categorize.NewService(mockProvider)
	chatService := chat.NewService(mockProvider)
	handler := NewHandler(categorizeService, chatService)

	reqBody := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Test"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/internal/categorize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Categorize(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestHandler_ChatQuery_Success(t *testing.T) {
	handler, _ := setupTestHandler()

	reqBody := domain.ChatQueryRequest{
		Question: "What did I spend?",
		LedgerSummary: domain.LedgerSummary{
			RangeStart:       "2024-01-01",
			RangeEnd:         "2024-01-31",
			TotalsByCategory: map[string]float64{},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/internal/chat/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ChatQuery(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusOK)
	}

	var resp domain.ChatQueryResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Answer != "Test answer" {
		t.Errorf("Answer: got %q, want %q", resp.Answer, "Test answer")
	}

	if len(resp.Citations) != 1 {
		t.Errorf("Citations count: got %d, want 1", len(resp.Citations))
	}
}

func TestHandler_ChatQuery_EmptyQuestion(t *testing.T) {
	handler, _ := setupTestHandler()

	reqBody := domain.ChatQueryRequest{
		Question: "",
		LedgerSummary: domain.LedgerSummary{
			RangeStart:       "2024-01-01",
			RangeEnd:         "2024-01-31",
			TotalsByCategory: map[string]float64{},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/internal/chat/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ChatQuery(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandler_ChatQuery_MissingDateRange(t *testing.T) {
	handler, _ := setupTestHandler()

	reqBody := domain.ChatQueryRequest{
		Question: "What did I spend?",
		LedgerSummary: domain.LedgerSummary{
			RangeStart:       "",
			RangeEnd:         "",
			TotalsByCategory: map[string]float64{},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/internal/chat/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ChatQuery(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandler_ChatQuery_MethodNotAllowed(t *testing.T) {
	handler, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/internal/chat/query", nil)
	rr := httptest.NewRecorder()

	handler.ChatQuery(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestRegisterRoutes(t *testing.T) {
	handler, _ := setupTestHandler()
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	tests := []struct {
		path       string
		method     string
		wantStatus int
	}{
		{"/health", http.MethodGet, http.StatusOK},
		{"/internal/categorize", http.MethodPost, http.StatusBadRequest}, // empty body = bad request
		{"/internal/chat/query", http.MethodPost, http.StatusBadRequest}, // empty body = bad request
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.method, func(t *testing.T) {
			var req *http.Request
			if tt.method == http.MethodPost {
				req = httptest.NewRequest(tt.method, tt.path, bytes.NewReader([]byte(`{}`)))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}
