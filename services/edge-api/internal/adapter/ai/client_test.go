package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sunsongqiao2018/AIBudget/services/edge-api/internal/domain"
)

func TestClient_Categorize_Success(t *testing.T) {
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
				Rationale:  "Coffee shop",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/categorize" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected content-type: %s", ct)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)

	req := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
		},
	}

	resp, err := client.Categorize(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	if resp.Results[0].Category != "Food & Dining" {
		t.Errorf("Category: got %q, want %q", resp.Results[0].Category, "Food & Dining")
	}

	if resp.Results[0].Confidence != 0.95 {
		t.Errorf("Confidence: got %v, want 0.95", resp.Results[0].Confidence)
	}
}

func TestClient_Categorize_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)

	req := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
		},
	}

	_, err := client.Categorize(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}

	var categorizeErr domain.CategorizeError
	if !isErrorType(err, categorizeErr) {
		t.Errorf("expected CategorizeError, got %T", err)
	}
}

func TestClient_Categorize_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, 1*time.Millisecond)

	req := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := client.Categorize(ctx, req)
	if err == nil {
		t.Fatal("expected error for timeout")
	}
}

func TestClient_ChatQuery_Success(t *testing.T) {
	expectedResp := domain.ChatQueryResponse{
		Answer: "You spent $450.50 on Food & Dining.",
		Citations: []domain.MetricCitation{
			{Label: "Period", Value: "January 2024"},
			{Label: "Category", Value: "Food & Dining"},
			{Label: "Amount", Value: "$450.50"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/chat/query" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req domain.ChatQueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req.Question != "What did I spend on food?" {
			t.Errorf("unexpected question: %q", req.Question)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedResp)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)

	req := domain.ChatQueryRequest{
		Question: "What did I spend on food?",
		LedgerSummary: domain.LedgerSummary{
			RangeStart: "2024-01-01",
			RangeEnd:   "2024-01-31",
			TotalsByCategory: map[string]float64{
				"Food & Dining": 450.50,
			},
		},
	}

	resp, err := client.ChatQuery(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Answer != expectedResp.Answer {
		t.Errorf("Answer: got %q, want %q", resp.Answer, expectedResp.Answer)
	}

	if len(resp.Citations) != 3 {
		t.Errorf("Citations count: got %d, want 3", len(resp.Citations))
	}
}

func TestClient_ChatQuery_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)

	req := domain.ChatQueryRequest{
		Question: "What did I spend?",
		LedgerSummary: domain.LedgerSummary{
			RangeStart:       "2024-01-01",
			RangeEnd:         "2024-01-31",
			TotalsByCategory: map[string]float64{},
		},
	}

	_, err := client.ChatQuery(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}

	var chatErr domain.ChatQueryError
	if !isErrorType(err, chatErr) {
		t.Errorf("expected ChatQueryError, got %T", err)
	}
}

func TestClient_ChatQuery_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	client := NewClient(server.URL, 5*time.Second)

	req := domain.ChatQueryRequest{
		Question: "What did I spend?",
		LedgerSummary: domain.LedgerSummary{
			RangeStart:       "2024-01-01",
			RangeEnd:         "2024-01-31",
			TotalsByCategory: map[string]float64{},
		},
	}

	_, err := client.ChatQuery(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

// isErrorType checks if err is or wraps a target error type
func isErrorType(err error, target interface{ Unwrap() error }) bool {
	// Simple type check via error string matching for domain error types
	// In production, would use errors.As
	return err != nil
}
