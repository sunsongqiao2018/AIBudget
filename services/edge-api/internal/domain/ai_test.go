package domain

import (
	"context"
	"errors"
	"testing"
)

// MockAIOrchestratorClient is a test implementation of AIOrchestratorClient.
type MockAIOrchestratorClient struct {
	CategorizeFunc func(ctx context.Context, req CategorizeRequest) (CategorizeResponse, error)
	ChatQueryFunc  func(ctx context.Context, req ChatQueryRequest) (ChatQueryResponse, error)
}

func (m *MockAIOrchestratorClient) Categorize(ctx context.Context, req CategorizeRequest) (CategorizeResponse, error) {
	if m.CategorizeFunc != nil {
		return m.CategorizeFunc(ctx, req)
	}
	return CategorizeResponse{}, nil
}

func (m *MockAIOrchestratorClient) ChatQuery(ctx context.Context, req ChatQueryRequest) (ChatQueryResponse, error) {
	if m.ChatQueryFunc != nil {
		return m.ChatQueryFunc(ctx, req)
	}
	return ChatQueryResponse{}, nil
}

func TestMockAIOrchestratorClient_Categorize(t *testing.T) {
	expectedResp := CategorizeResponse{
		Results: []CategorizedTransaction{
			{
				TransactionCandidate: TransactionCandidate{
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

	mock := &MockAIOrchestratorClient{
		CategorizeFunc: func(ctx context.Context, req CategorizeRequest) (CategorizeResponse, error) {
			if len(req.Transactions) == 0 {
				return CategorizeResponse{}, errors.New("no transactions provided")
			}
			return expectedResp, nil
		},
	}

	req := CategorizeRequest{
		Transactions: []TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
		},
	}

	resp, err := mock.Categorize(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	if resp.Results[0].Category != "Food & Dining" {
		t.Errorf("Category: got %q, want %q", resp.Results[0].Category, "Food & Dining")
	}

	// Test error case
	_, err = mock.Categorize(context.Background(), CategorizeRequest{})
	if err == nil {
		t.Error("expected error for empty transactions")
	}
}

func TestMockAIOrchestratorClient_ChatQuery(t *testing.T) {
	expectedResp := ChatQueryResponse{
		Answer: "You spent $450.50 on Food & Dining in January.",
		Citations: []MetricCitation{
			{Label: "Period", Value: "2024-01-01 to 2024-01-31"},
			{Label: "Category Total", Value: "$450.50"},
		},
	}

	mock := &MockAIOrchestratorClient{
		ChatQueryFunc: func(ctx context.Context, req ChatQueryRequest) (ChatQueryResponse, error) {
			if req.Question == "" {
				return ChatQueryResponse{}, errors.New("empty question")
			}
			return expectedResp, nil
		},
	}

	req := ChatQueryRequest{
		Question: "What did I spend on food?",
		LedgerSummary: LedgerSummary{
			RangeStart: "2024-01-01",
			RangeEnd:   "2024-01-31",
			TotalsByCategory: map[string]float64{
				"Food & Dining": 450.50,
			},
		},
	}

	resp, err := mock.ChatQuery(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Answer != expectedResp.Answer {
		t.Errorf("Answer: got %q, want %q", resp.Answer, expectedResp.Answer)
	}

	if len(resp.Citations) != 2 {
		t.Errorf("Citations count: got %d, want 2", len(resp.Citations))
	}

	// Test error case
	_, err = mock.ChatQuery(context.Background(), ChatQueryRequest{})
	if err == nil {
		t.Error("expected error for empty question")
	}
}

func TestCategorizeError(t *testing.T) {
	underlying := errors.New("network timeout")
	err := CategorizeError{
		Message: "failed to reach AI orchestrator",
		Cause:   underlying,
	}

	want := "categorize failed: failed to reach AI orchestrator"
	if err.Error() != want {
		t.Errorf("Error(): got %q, want %q", err.Error(), want)
	}

	if err.Unwrap() != underlying {
		t.Error("Unwrap() did not return underlying error")
	}
}

func TestChatQueryError(t *testing.T) {
	underlying := errors.New("invalid response format")
	err := ChatQueryError{
		Message: "AI service returned malformed data",
		Cause:   underlying,
	}

	want := "chat query failed: AI service returned malformed data"
	if err.Error() != want {
		t.Errorf("Error(): got %q, want %q", err.Error(), want)
	}

	if err.Unwrap() != underlying {
		t.Error("Unwrap() did not return underlying error")
	}
}
