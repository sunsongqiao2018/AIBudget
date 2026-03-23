// Package categorize provides the categorization application service.
package categorize

import (
	"context"
	"fmt"

	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/domain"
)

// Service handles transaction categorization use cases.
type Service struct {
	llmProvider domain.LLMProvider
}

// NewService creates a new categorization service.
func NewService(llmProvider domain.LLMProvider) *Service {
	return &Service{
		llmProvider: llmProvider,
	}
}

// Categorize processes transaction candidates and returns categorized results.
// It validates input, calls the LLM provider, and returns the results.
func (s *Service) Categorize(ctx context.Context, req domain.CategorizeRequest) (domain.CategorizeResponse, error) {
	// Validate request
	if len(req.Transactions) == 0 {
		return domain.CategorizeResponse{}, &domain.LLMError{
			Op:      "categorize",
			Message: "no transactions provided",
		}
	}

	// Validate each transaction
	for i, tx := range req.Transactions {
		if err := validateTransaction(tx, i); err != nil {
			return domain.CategorizeResponse{}, &domain.LLMError{
				Op:      "categorize",
				Message: err.Error(),
			}
		}
	}

	// Call LLM provider
	results, err := s.llmProvider.Categorize(ctx, req.Transactions)
	if err != nil {
		return domain.CategorizeResponse{}, &domain.LLMError{
			Op:      "categorize",
			Message: "LLM provider failed",
			Err:     err,
		}
	}

	// Validate results count matches input count
	if len(results) != len(req.Transactions) {
		return domain.CategorizeResponse{}, &domain.LLMError{
			Op:      "categorize",
			Message: fmt.Sprintf("result count mismatch: expected %d, got %d", len(req.Transactions), len(results)),
		}
	}

	return domain.CategorizeResponse{Results: results}, nil
}

// validateTransaction validates a single transaction candidate.
func validateTransaction(tx domain.TransactionCandidate, index int) error {
	if tx.Date == "" {
		return fmt.Errorf("transaction[%d]: date is required", index)
	}
	if tx.Merchant == "" {
		return fmt.Errorf("transaction[%d]: merchant is required", index)
	}
	if tx.Currency == "" {
		return fmt.Errorf("transaction[%d]: currency is required", index)
	}
	if tx.Amount <= 0 {
		return fmt.Errorf("transaction[%d]: amount must be positive", index)
	}
	if len(tx.Currency) != 3 {
		return fmt.Errorf("transaction[%d]: currency must be 3 characters", index)
	}
	return nil
}
