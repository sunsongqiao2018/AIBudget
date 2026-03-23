package domain

import (
	"context"
	"errors"
	"testing"
)

// MockLLMProvider is a test implementation of LLMProvider.
type MockLLMProvider struct {
	CategorizeFunc func(ctx context.Context, candidates []TransactionCandidate) ([]CategorizedTransaction, error)
	ChatQueryFunc  func(ctx context.Context, question string, ledger LedgerSummary) (ChatResponse, error)
}

func (m *MockLLMProvider) Categorize(ctx context.Context, candidates []TransactionCandidate) ([]CategorizedTransaction, error) {
	if m.CategorizeFunc != nil {
		return m.CategorizeFunc(ctx, candidates)
	}
	return nil, nil
}

func (m *MockLLMProvider) ChatQuery(ctx context.Context, question string, ledger LedgerSummary) (ChatResponse, error) {
	if m.ChatQueryFunc != nil {
		return m.ChatQueryFunc(ctx, question, ledger)
	}
	return ChatResponse{}, nil
}

func TestMockLLMProvider_Categorize(t *testing.T) {
	expectedResults := []CategorizedTransaction{
		{
			TransactionCandidate: TransactionCandidate{
				Date:     "2024-01-15",
				Amount:   45.67,
				Currency: "USD",
				Merchant: "Starbucks",
			},
			Category:   "Food & Dining",
			Confidence: 0.95,
			Rationale:  "Coffee shop purchase",
		},
		{
			TransactionCandidate: TransactionCandidate{
				Date:     "2024-01-16",
				Amount:   23.50,
				Currency: "USD",
				Merchant: "Uber",
			},
			Category:   "Transportation",
			Confidence: 0.88,
			Rationale:  "Ride sharing service",
		},
	}

	mock := &MockLLMProvider{
		CategorizeFunc: func(ctx context.Context, candidates []TransactionCandidate) ([]CategorizedTransaction, error) {
			if len(candidates) == 0 {
				return nil, ErrCategorizationFailed
			}
			return expectedResults, nil
		},
	}

	candidates := []TransactionCandidate{
		{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
		{Date: "2024-01-16", Amount: 23.50, Currency: "USD", Merchant: "Uber"},
	}

	results, err := mock.Categorize(context.Background(), candidates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Category != "Food & Dining" {
		t.Errorf("Category[0]: got %q, want %q", results[0].Category, "Food & Dining")
	}

	if results[1].Confidence != 0.88 {
		t.Errorf("Confidence[1]: got %v, want 0.88", results[1].Confidence)
	}

	// Test error case
	_, err = mock.Categorize(context.Background(), nil)
	if !errors.Is(err, ErrCategorizationFailed) {
		t.Errorf("expected ErrCategorizationFailed, got %v", err)
	}
}

func TestMockLLMProvider_ChatQuery(t *testing.T) {
	expectedResp := ChatResponse{
		Answer: "You spent $450.50 on Food & Dining in January 2024.",
		Citations: []MetricCitation{
			{Label: "Period", Value: "January 2024"},
			{Label: "Category", Value: "Food & Dining"},
			{Label: "Amount", Value: "$450.50"},
		},
	}

	mock := &MockLLMProvider{
		ChatQueryFunc: func(ctx context.Context, question string, ledger LedgerSummary) (ChatResponse, error) {
			if question == "" {
				return ChatResponse{}, ErrChatQueryFailed
			}
			return expectedResp, nil
		},
	}

	ledger := LedgerSummary{
		RangeStart: "2024-01-01",
		RangeEnd:   "2024-01-31",
		TotalsByCategory: map[string]float64{
			"Food & Dining": 450.50,
		},
	}

	resp, err := mock.ChatQuery(context.Background(), "What did I spend on food?", ledger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Answer != expectedResp.Answer {
		t.Errorf("Answer: got %q, want %q", resp.Answer, expectedResp.Answer)
	}

	if len(resp.Citations) != 3 {
		t.Errorf("Citations count: got %d, want 3", len(resp.Citations))
	}

	// Test error case
	_, err = mock.ChatQuery(context.Background(), "", ledger)
	if !errors.Is(err, ErrChatQueryFailed) {
		t.Errorf("expected ErrChatQueryFailed, got %v", err)
	}
}

func TestDefaultCategories(t *testing.T) {
	cats := DefaultCategories()

	expected := []string{
		"Food & Dining",
		"Transportation",
		"Shopping",
		"Entertainment",
		"Bills & Utilities",
		"Health & Wellness",
		"Travel",
		"Education",
		"Other",
	}

	if len(cats) != len(expected) {
		t.Fatalf("expected %d categories, got %d", len(expected), len(cats))
	}

	for i, cat := range expected {
		if cats[i] != cat {
			t.Errorf("category[%d]: got %q, want %q", i, cats[i], cat)
		}
	}
}

func TestLLMError(t *testing.T) {
	underlying := errors.New("connection refused")
	err := &LLMError{
		Op:      "categorize",
		Message: "failed to connect to provider",
		Err:     underlying,
	}

	want := "llm categorize failed: failed to connect to provider"
	if err.Error() != want {
		t.Errorf("Error(): got %q, want %q", err.Error(), want)
	}

	if err.Unwrap() != underlying {
		t.Error("Unwrap() did not return underlying error")
	}
}

func TestProviderConfig_Values(t *testing.T) {
	config := ProviderConfig{
		Provider:    "openai",
		APIKey:      "test-key",
		Model:       "gpt-4",
		TimeoutSecs: 30,
		MaxRetries:  3,
	}

	if config.Provider != "openai" {
		t.Errorf("Provider: got %q, want %q", config.Provider, "openai")
	}

	if config.TimeoutSecs != 30 {
		t.Errorf("TimeoutSecs: got %d, want 30", config.TimeoutSecs)
	}

	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries: got %d, want 3", config.MaxRetries)
	}
}

func TestCategorizationConfig(t *testing.T) {
	config := CategorizationConfig{
		Categories: DefaultCategories(),
	}

	if len(config.Categories) != 9 {
		t.Errorf("expected 9 categories, got %d", len(config.Categories))
	}
}

func TestErrors_AreDistinct(t *testing.T) {
	errs := []error{
		ErrCategorizationFailed,
		ErrChatQueryFailed,
		ErrProviderTimeout,
		ErrProviderUnavailable,
	}

	seen := make(map[string]bool)
	for _, err := range errs {
		if seen[err.Error()] {
			t.Errorf("duplicate error message: %q", err.Error())
		}
		seen[err.Error()] = true
	}
}
