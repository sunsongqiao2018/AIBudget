// Package domain defines LLM-related domain interfaces for the ai-orchestrator service.
package domain

import (
	"context"
	"errors"
)

// Common errors for LLM operations.
var (
	ErrCategorizationFailed = errors.New("categorization failed")
	ErrChatQueryFailed      = errors.New("chat query failed")
	ErrProviderTimeout      = errors.New("LLM provider timeout")
	ErrProviderUnavailable  = errors.New("LLM provider unavailable")
)

// LLMProvider defines the interface for LLM provider implementations.
// This abstraction allows swapping between mock, OpenAI, or other providers.
type LLMProvider interface {
	// Categorize analyzes transaction candidates and returns suggested
	// categories with confidence scores.
	Categorize(ctx context.Context, candidates []TransactionCandidate) ([]CategorizedTransaction, error)

	// ChatQuery answers a user question using the provided ledger summary.
	// Returns a natural language answer with cited metrics.
	ChatQuery(ctx context.Context, question string, ledger LedgerSummary) (ChatResponse, error)
}

// ChatResponse contains the answer and citations from a chat query.
type ChatResponse struct {
	Answer    string
	Citations []MetricCitation
}

// ProviderConfig holds configuration for LLM providers.
type ProviderConfig struct {
	Provider    string  // "mock", "openai"
	APIKey      string
	Model       string
	TimeoutSecs int
	MaxRetries  int
}

// CategorizationConfig defines category options for the LLM.
type CategorizationConfig struct {
	Categories []string
}

// DefaultCategories returns the standard spending categories.
func DefaultCategories() []string {
	return []string{
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
}

// LLMError wraps errors from the LLM provider with context.
type LLMError struct {
	Op      string // operation that failed ("categorize", "chat")
	Message string
	Err     error
}

func (e *LLMError) Error() string {
	return "llm " + e.Op + " failed: " + e.Message
}

func (e *LLMError) Unwrap() error {
	return e.Err
}
