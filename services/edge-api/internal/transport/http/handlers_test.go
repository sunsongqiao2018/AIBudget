package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sunsongqiao2018/AIBudget/services/edge-api/internal/domain"
)

// mockAIClient is a test implementation of domain.AIOrchestratorClient.
type mockAIClient struct {
	categorizeFunc func(ctx context.Context, req domain.CategorizeRequest) (domain.CategorizeResponse, error)
	chatQueryFunc  func(ctx context.Context, req domain.ChatQueryRequest) (domain.ChatQueryResponse, error)
}

func (m *mockAIClient) Categorize(ctx context.Context, req domain.CategorizeRequest) (domain.CategorizeResponse, error) {
	if m.categorizeFunc != nil {
		return m.categorizeFunc(ctx, req)
	}
	return domain.CategorizeResponse{}, nil
}

func (m *mockAIClient) ChatQuery(ctx context.Context, req domain.ChatQueryRequest) (domain.ChatQueryResponse, error) {
	if m.chatQueryFunc != nil {
		return m.chatQueryFunc(ctx, req)
	}
	return domain.ChatQueryResponse{}, nil
}

func TestHandler_Health(t *testing.T) {
	handler := NewHandler(&mockAIClient{})

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

	if resp["service"] != "edge-api" {
		t.Errorf("service: got %q, want %q", resp["service"], "edge-api")
	}
}

func TestHandler_Health_MethodNotAllowed(t *testing.T) {
	handler := NewHandler(&mockAIClient{})

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()

	handler.Health(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandler_Categorize_Success(t *testing.T) {
	expectedResp := domain.CategorizeResponse{
		Results: []domain.CategorizedTransaction{
			{
				TransactionCandidate: domain.TransactionCandidate{
					Date:     "2024-01-15",
					Amount:   45.67,
					Currency: "USD",
					Merchant: "Starbucks",
				},
				Category:   "Food & Dining",
				Confidence: 0.95,
			},
		},
	}

	mockClient := &mockAIClient{
		categorizeFunc: func(ctx context.Context, req domain.CategorizeRequest) (domain.CategorizeResponse, error) {
			if len(req.Transactions) != 1 {
				t.Errorf("expected 1 transaction, got %d", len(req.Transactions))
			}
			return expectedResp, nil
		},
	}

	handler := NewHandler(mockClient)

	reqBody := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/ai/categorize", bytes.NewReader(body))
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

	if resp.Results[0].Category != "Food & Dining" {
		t.Errorf("Category: got %q, want %q", resp.Results[0].Category, "Food & Dining")
	}
}

func TestHandler_Categorize_EmptyTransactions(t *testing.T) {
	handler := NewHandler(&mockAIClient{})

	reqBody := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/ai/categorize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Categorize(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}

	var resp domain.ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if resp.Error == "" {
		t.Error("expected error message in response")
	}
}

func TestHandler_Categorize_InvalidJSON(t *testing.T) {
	handler := NewHandler(&mockAIClient{})

	req := httptest.NewRequest(http.MethodPost, "/v1/ai/categorize", bytes.NewReader([]byte(`invalid`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Categorize(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandler_Categorize_WrongContentType(t *testing.T) {
	handler := NewHandler(&mockAIClient{})

	req := httptest.NewRequest(http.MethodPost, "/v1/ai/categorize", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	handler.Categorize(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandler_Categorize_MethodNotAllowed(t *testing.T) {
	handler := NewHandler(&mockAIClient{})

	req := httptest.NewRequest(http.MethodGet, "/v1/ai/categorize", nil)
	rr := httptest.NewRecorder()

	handler.Categorize(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandler_Categorize_MissingRequiredFields(t *testing.T) {
	handler := NewHandler(&mockAIClient{})

	tests := []struct {
		name        string
		transaction domain.TransactionCandidate
		wantErr     string
	}{
		{
			name:        "missing date",
			transaction: domain.TransactionCandidate{Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
			wantErr:     "date is required",
		},
		{
			name:        "missing currency",
			transaction: domain.TransactionCandidate{Date: "2024-01-15", Amount: 45.67, Merchant: "Starbucks"},
			wantErr:     "currency is required",
		},
		{
			name:        "missing merchant",
			transaction: domain.TransactionCandidate{Date: "2024-01-15", Amount: 45.67, Currency: "USD"},
			wantErr:     "merchant is required",
		},
		{
			name:        "zero amount",
			transaction: domain.TransactionCandidate{Date: "2024-01-15", Amount: 0, Currency: "USD", Merchant: "Starbucks"},
			wantErr:     "amount must be positive",
		},
		{
			name:        "negative amount",
			transaction: domain.TransactionCandidate{Date: "2024-01-15", Amount: -10, Currency: "USD", Merchant: "Starbucks"},
			wantErr:     "amount must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := domain.CategorizeRequest{
				Transactions: []domain.TransactionCandidate{tt.transaction},
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/v1/ai/categorize", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Categorize(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandler_Categorize_AIClientError(t *testing.T) {
	mockClient := &mockAIClient{
		categorizeFunc: func(ctx context.Context, req domain.CategorizeRequest) (domain.CategorizeResponse, error) {
			return domain.CategorizeResponse{}, domain.CategorizeError{
				Message: "AI service unavailable",
				Cause:   errors.New("connection refused"),
			}
		},
	}

	handler := NewHandler(mockClient)

	reqBody := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/ai/categorize", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.Categorize(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadGateway)
	}
}

func TestHandler_ChatQuery_Success(t *testing.T) {
	expectedResp := domain.ChatQueryResponse{
		Answer: "You spent $450.50 on Food & Dining.",
		Citations: []domain.MetricCitation{
			{Label: "Period", Value: "January 2024"},
			{Label: "Amount", Value: "$450.50"},
		},
	}

	mockClient := &mockAIClient{
		chatQueryFunc: func(ctx context.Context, req domain.ChatQueryRequest) (domain.ChatQueryResponse, error) {
			if req.Question != "What did I spend on food?" {
				t.Errorf("question: got %q, want %q", req.Question, "What did I spend on food?")
			}
			return expectedResp, nil
		},
	}

	handler := NewHandler(mockClient)

	reqBody := domain.ChatQueryRequest{
		Question: "What did I spend on food?",
		LedgerSummary: domain.LedgerSummary{
			RangeStart: "2024-01-01",
			RangeEnd:   "2024-01-31",
			TotalsByCategory: map[string]float64{
				"Food & Dining": 450.50,
			},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/ai/chat/query", bytes.NewReader(body))
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

	if resp.Answer != expectedResp.Answer {
		t.Errorf("Answer: got %q, want %q", resp.Answer, expectedResp.Answer)
	}

	if len(resp.Citations) != 2 {
		t.Errorf("Citations: got %d, want 2", len(resp.Citations))
	}
}

func TestHandler_ChatQuery_EmptyQuestion(t *testing.T) {
	handler := NewHandler(&mockAIClient{})

	reqBody := domain.ChatQueryRequest{
		Question: "",
		LedgerSummary: domain.LedgerSummary{
			RangeStart:       "2024-01-01",
			RangeEnd:         "2024-01-31",
			TotalsByCategory: map[string]float64{},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/ai/chat/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ChatQuery(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandler_ChatQuery_MissingDateRange(t *testing.T) {
	handler := NewHandler(&mockAIClient{})

	reqBody := domain.ChatQueryRequest{
		Question: "What did I spend?",
		LedgerSummary: domain.LedgerSummary{
			RangeStart:       "",
			RangeEnd:         "",
			TotalsByCategory: map[string]float64{},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/ai/chat/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ChatQuery(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandler_ChatQuery_MethodNotAllowed(t *testing.T) {
	handler := NewHandler(&mockAIClient{})

	req := httptest.NewRequest(http.MethodGet, "/v1/ai/chat/query", nil)
	rr := httptest.NewRecorder()

	handler.ChatQuery(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandler_ChatQuery_AIClientError(t *testing.T) {
	mockClient := &mockAIClient{
		chatQueryFunc: func(ctx context.Context, req domain.ChatQueryRequest) (domain.ChatQueryResponse, error) {
			return domain.ChatQueryResponse{}, domain.ChatQueryError{
				Message: "AI service timeout",
				Cause:   context.DeadlineExceeded,
			}
		},
	}

	handler := NewHandler(mockClient)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	reqBody := domain.ChatQueryRequest{
		Question: "What did I spend?",
		LedgerSummary: domain.LedgerSummary{
			RangeStart:       "2024-01-01",
			RangeEnd:         "2024-01-31",
			TotalsByCategory: map[string]float64{},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/ai/chat/query", bytes.NewReader(body)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ChatQuery(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("status: got %d, want %d", rr.Code, http.StatusBadGateway)
	}
}

func TestRegisterRoutes(t *testing.T) {
	mockClient := &mockAIClient{
		categorizeFunc: func(ctx context.Context, req domain.CategorizeRequest) (domain.CategorizeResponse, error) {
			return domain.CategorizeResponse{Results: []domain.CategorizedTransaction{}}, nil
		},
	}

	handler := NewHandler(mockClient)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Test that routes are registered by making requests
	tests := []struct {
		path       string
		method     string
		wantStatus int
	}{
		{"/health", http.MethodGet, http.StatusOK},
		{"/v1/ai/categorize", http.MethodPost, http.StatusBadRequest}, // empty body = bad request
		{"/v1/ai/chat/query", http.MethodPost, http.StatusBadRequest}, // empty body = bad request
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
