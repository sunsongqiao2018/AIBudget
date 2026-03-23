// Package chat provides the chat query application service.
package chat

import (
	"context"
	"fmt"

	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/domain"
)

// Service handles chat query use cases.
type Service struct {
	llmProvider domain.LLMProvider
}

// NewService creates a new chat service.
func NewService(llmProvider domain.LLMProvider) *Service {
	return &Service{
		llmProvider: llmProvider,
	}
}

// Query processes a chat query and returns an answer with citations.
// It validates input, calls the LLM provider, and returns the response.
func (s *Service) Query(ctx context.Context, req domain.ChatQueryRequest) (domain.ChatQueryResponse, error) {
	// Validate request
	if err := validateRequest(req); err != nil {
		return domain.ChatQueryResponse{}, &domain.LLMError{
			Op:      "chat",
			Message: err.Error(),
		}
	}

	// Call LLM provider
	resp, err := s.llmProvider.ChatQuery(ctx, req.Question, req.LedgerSummary)
	if err != nil {
		return domain.ChatQueryResponse{}, &domain.LLMError{
			Op:      "chat",
			Message: "LLM provider failed",
			Err:     err,
		}
	}

	// Validate response
	if resp.Answer == "" {
		return domain.ChatQueryResponse{}, &domain.LLMError{
			Op:      "chat",
			Message: "LLM returned empty answer",
		}
	}

	return domain.ChatQueryResponse{
		Answer:    resp.Answer,
		Citations:   resp.Citations,
	}, nil
}

// validateRequest validates the chat query request.
func validateRequest(req domain.ChatQueryRequest) error {
	if req.Question == "" {
		return fmt.Errorf("question is required")
	}

	if len(req.Question) < 3 {
		return fmt.Errorf("question must be at least 3 characters")
	}

	if req.LedgerSummary.RangeStart == "" {
		return fmt.Errorf("ledgerSummary.rangeStart is required")
	}

	if req.LedgerSummary.RangeEnd == "" {
		return fmt.Errorf("ledgerSummary.rangeEnd is required")
	}

	// TotalsByCategory can be empty (no spending in period is valid)
	return nil
}
