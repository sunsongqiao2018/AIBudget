package categorize

import (
	"context"
	"errors"
	"testing"

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

func TestService_Categorize_Success(t *testing.T) {
	expectedResults := []domain.CategorizedTransaction{
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
	}

	mockProvider := &mockLLMProvider{
		categorizeFunc: func(ctx context.Context, candidates []domain.TransactionCandidate) ([]domain.CategorizedTransaction, error) {
			if len(candidates) != 1 {
				t.Errorf("expected 1 candidate, got %d", len(candidates))
			}
			return expectedResults, nil
		},
	}

	service := NewService(mockProvider)

	req := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
		},
	}

	resp, err := service.Categorize(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}

	if resp.Results[0].Category != "Food & Dining" {
		t.Errorf("Category: got %q, want %q", resp.Results[0].Category, "Food & Dining")
	}
}

func TestService_Categorize_EmptyTransactions(t *testing.T) {
	service := NewService(&mockLLMProvider{})

	req := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{},
	}

	_, err := service.Categorize(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty transactions")
	}

	var llmErr *domain.LLMError
	if !errors.As(err, &llmErr) {
		t.Errorf("expected *domain.LLMError, got %T", err)
	}
}

func TestService_Categorize_InvalidTransaction(t *testing.T) {
	service := NewService(&mockLLMProvider{})

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
			name:        "missing merchant",
			transaction: domain.TransactionCandidate{Date: "2024-01-15", Amount: 45.67, Currency: "USD"},
			wantErr:     "merchant is required",
		},
		{
			name:        "missing currency",
			transaction: domain.TransactionCandidate{Date: "2024-01-15", Amount: 45.67, Merchant: "Starbucks"},
			wantErr:     "currency is required",
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
		{
			name:        "invalid currency length",
			transaction: domain.TransactionCandidate{Date: "2024-01-15", Amount: 45.67, Currency: "US", Merchant: "Starbucks"},
			wantErr:     "currency must be 3 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := domain.CategorizeRequest{
				Transactions: []domain.TransactionCandidate{tt.transaction},
			}

			_, err := service.Categorize(context.Background(), req)
			if err == nil {
				t.Fatal("expected error")
			}

			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("error message: got %q, should contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestService_Categorize_ProviderError(t *testing.T) {
	mockProvider := &mockLLMProvider{
		categorizeFunc: func(ctx context.Context, candidates []domain.TransactionCandidate) ([]domain.CategorizedTransaction, error) {
			return nil, errors.New("LLM service unavailable")
		},
	}

	service := NewService(mockProvider)

	req := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
		},
	}

	_, err := service.Categorize(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for provider failure")
	}

	var llmErr *domain.LLMError
	if !errors.As(err, &llmErr) {
		t.Errorf("expected *domain.LLMError, got %T", err)
	}
}

func TestService_Categorize_ResultCountMismatch(t *testing.T) {
	mockProvider := &mockLLMProvider{
		categorizeFunc: func(ctx context.Context, candidates []domain.TransactionCandidate) ([]domain.CategorizedTransaction, error) {
			// Return fewer results than input
			return []domain.CategorizedTransaction{
				{TransactionCandidate: candidates[0], Category: "Food", Confidence: 0.9},
			}, nil
		},
	}

	service := NewService(mockProvider)

	req := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
			{Date: "2024-01-16", Amount: 23.50, Currency: "USD", Merchant: "Uber"},
		},
	}

	_, err := service.Categorize(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for result count mismatch")
	}

	if !contains(err.Error(), "result count mismatch") {
		t.Errorf("error should mention 'result count mismatch', got: %s", err.Error())
	}
}

func TestService_Categorize_MultipleTransactions(t *testing.T) {
	expectedResults := []domain.CategorizedTransaction{
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
		{
			TransactionCandidate: domain.TransactionCandidate{
				Date:     "2024-01-16",
				Amount:   23.50,
				Currency: "USD",
				Merchant: "Uber",
			},
			Category:   "Transportation",
			Confidence: 0.88,
		},
	}

	mockProvider := &mockLLMProvider{
		categorizeFunc: func(ctx context.Context, candidates []domain.TransactionCandidate) ([]domain.CategorizedTransaction, error) {
			return expectedResults, nil
		},
	}

	service := NewService(mockProvider)

	req := domain.CategorizeRequest{
		Transactions: []domain.TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
			{Date: "2024-01-16", Amount: 23.50, Currency: "USD", Merchant: "Uber"},
		},
	}

	resp, err := service.Categorize(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Results))
	}

	if resp.Results[0].Merchant != "Starbucks" {
		t.Errorf("first result merchant: got %q, want %q", resp.Results[0].Merchant, "Starbucks")
	}

	if resp.Results[1].Merchant != "Uber" {
		t.Errorf("second result merchant: got %q, want %q", resp.Results[1].Merchant, "Uber")
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(substr) > 0 && containsInternal(s, substr)))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
