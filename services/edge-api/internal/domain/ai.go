// Package domain defines AI-related domain interfaces for the edge-api service.
package domain

import (
	"context"
)

// AIOrchestratorClient defines the interface for communicating with the
// ai-orchestrator service. This abstraction allows for easy mocking in tests.
type AIOrchestratorClient interface {
	// Categorize sends transaction candidates to the AI orchestrator
	// for automatic categorization. Returns categorized transactions with
	// confidence scores.
	Categorize(ctx context.Context, req CategorizeRequest) (CategorizeResponse, error)

	// ChatQuery sends a user question with ledger context to the AI orchestrator
	// for a natural language response with cited metrics.
	ChatQuery(ctx context.Context, req ChatQueryRequest) (ChatQueryResponse, error)
}

// CategorizeError represents errors from the categorization flow.
type CategorizeError struct {
	Message string
	Cause   error
}

func (e CategorizeError) Error() string {
	return "categorize failed: " + e.Message
}

func (e CategorizeError) Unwrap() error {
	return e.Cause
}

// ChatQueryError represents errors from the chat query flow.
type ChatQueryError struct {
	Message string
	Cause   error
}

func (e ChatQueryError) Error() string {
	return "chat query failed: " + e.Message
}

func (e ChatQueryError) Unwrap() error {
	return e.Cause
}
