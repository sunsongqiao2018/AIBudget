package chat

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

func TestService_Query_Success(t *testing.T) {
	expectedResp := domain.ChatResponse{
		Answer: "You spent $450.50 on Food & Dining.",
		Citations: []domain.MetricCitation{
			{Label: "Category", Value: "Food & Dining"},
			{Label: "Amount", Value: "$450.50"},
		},
	}

	mockProvider := &mockLLMProvider{
		chatQueryFunc: func(ctx context.Context, question string, ledger domain.LedgerSummary) (domain.ChatResponse, error) {
			if question != "What did I spend on food?" {
				t.Errorf("question: got %q, want %q", question, "What did I spend on food?")
			}
			return expectedResp, nil
		},
	}

	service := NewService(mockProvider)

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

	resp, err := service.Query(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Answer != expectedResp.Answer {
		t.Errorf("Answer: got %q, want %q", resp.Answer, expectedResp.Answer)
	}

	if len(resp.Citations) != 2 {
		t.Errorf("Citations count: got %d, want 2", len(resp.Citations))
	}
}

func TestService_Query_EmptyQuestion(t *testing.T) {
	service := NewService(&mockLLMProvider{})

	req := domain.ChatQueryRequest{
		Question: "",
		LedgerSummary: domain.LedgerSummary{
			RangeStart: "2024-01-01",
			RangeEnd:   "2024-01-31",
		},
	}

	_, err := service.Query(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty question")
	}

	if !contains(err.Error(), "question is required") {
		t.Errorf("error should mention 'question is required', got: %s", err.Error())
	}
}

func TestService_Query_ShortQuestion(t *testing.T) {
	service := NewService(&mockLLMProvider{})

	req := domain.ChatQueryRequest{
		Question: "Hi",
		LedgerSummary: domain.LedgerSummary{
			RangeStart: "2024-01-01",
			RangeEnd:   "2024-01-31",
		},
	}

	_, err := service.Query(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for short question")
	}

	if !contains(err.Error(), "at least 3 characters") {
		t.Errorf("error should mention 'at least 3 characters', got: %s", err.Error())
	}
}

func TestService_Query_MissingDateRange(t *testing.T) {
	service := NewService(&mockLLMProvider{})

	tests := []struct {
		name    string
		req     domain.ChatQueryRequest
		wantErr string
	}{
		{
			name: "missing rangeStart",
			req: domain.ChatQueryRequest{
				Question: "What did I spend?",
				LedgerSummary: domain.LedgerSummary{
					RangeStart:       "",
					RangeEnd:         "2024-01-31",
					TotalsByCategory: map[string]float64{},
				},
			},
			wantErr: "rangeStart is required",
		},
		{
			name: "missing rangeEnd",
			req: domain.ChatQueryRequest{
				Question: "What did I spend?",
				LedgerSummary: domain.LedgerSummary{
					RangeStart:       "2024-01-01",
					RangeEnd:         "",
					TotalsByCategory: map[string]float64{},
				},
			},
			wantErr: "rangeEnd is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Query(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error")
			}

			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("error should contain %q, got: %s", tt.wantErr, err.Error())
			}
		})
	}
}

func TestService_Query_EmptyCategoriesAllowed(t *testing.T) {
	expectedResp := domain.ChatResponse{
		Answer: "You had no spending in this period.",
		Citations: []domain.MetricCitation{
			{Label: "Period", Value: "January 2024"},
		},
	}

	mockProvider := &mockLLMProvider{
		chatQueryFunc: func(ctx context.Context, question string, ledger domain.LedgerSummary) (domain.ChatResponse, error) {
			return expectedResp, nil
		},
	}

	service := NewService(mockProvider)

	req := domain.ChatQueryRequest{
		Question: "What did I spend?",
		LedgerSummary: domain.LedgerSummary{
			RangeStart:       "2024-01-01",
			RangeEnd:         "2024-01-31",
			TotalsByCategory: map[string]float64{}, // Empty is valid
		},
	}

	resp, err := service.Query(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Answer != expectedResp.Answer {
		t.Errorf("Answer: got %q, want %q", resp.Answer, expectedResp.Answer)
	}
}

func TestService_Query_ProviderError(t *testing.T) {
	mockProvider := &mockLLMProvider{
		chatQueryFunc: func(ctx context.Context, question string, ledger domain.LedgerSummary) (domain.ChatResponse, error) {
			return domain.ChatResponse{}, errors.New("LLM service timeout")
		},
	}

	service := NewService(mockProvider)

	req := domain.ChatQueryRequest{
		Question: "What did I spend?",
		LedgerSummary: domain.LedgerSummary{
			RangeStart:       "2024-01-01",
			RangeEnd:         "2024-01-31",
			TotalsByCategory: map[string]float64{},
		},
	}

	_, err := service.Query(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for provider failure")
	}

	var llmErr *domain.LLMError
	if !errors.As(err, &llmErr) {
		t.Errorf("expected *domain.LLMError, got %T", err)
	}
}

func TestService_Query_EmptyAnswer(t *testing.T) {
	mockProvider := &mockLLMProvider{
		chatQueryFunc: func(ctx context.Context, question string, ledger domain.LedgerSummary) (domain.ChatResponse, error) {
			return domain.ChatResponse{
				Answer:    "", // Empty answer
				Citations: []domain.MetricCitation{},
			}, nil
		},
	}

	service := NewService(mockProvider)

	req := domain.ChatQueryRequest{
		Question: "What did I spend?",
		LedgerSummary: domain.LedgerSummary{
			RangeStart:       "2024-01-01",
			RangeEnd:         "2024-01-31",
			TotalsByCategory: map[string]float64{},
		},
	}

	_, err := service.Query(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty answer")
	}

	if !contains(err.Error(), "empty answer") {
		t.Errorf("error should mention 'empty answer', got: %s", err.Error())
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
