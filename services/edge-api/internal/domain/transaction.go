// Package domain defines core domain entities for the edge-api service.
package domain

import (
	"time"
)

// Transaction represents a spending transaction in the system.
type Transaction struct {
	ID       string
	Date     time.Time
	Amount   float64
	Currency string
	Merchant string
	Category string
	Note     string
}

// TransactionCandidate represents a raw transaction extracted from OCR
// before categorization.
type TransactionCandidate struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Merchant string  `json:"merchant"`
	Note     string  `json:"note,omitempty"`
}

// CategorizedTransaction extends TransactionCandidate with categorization results.
type CategorizedTransaction struct {
	TransactionCandidate
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	Rationale  string  `json:"rationale,omitempty"`
}

// CategorizeRequest is the input for the categorize endpoint.
type CategorizeRequest struct {
	Transactions []TransactionCandidate `json:"transactions"`
}

// CategorizeResponse is the output from the categorize endpoint.
type CategorizeResponse struct {
	Results []CategorizedTransaction `json:"results"`
}

// LedgerSummary provides spending context for chat queries.
type LedgerSummary struct {
	RangeStart       string             `json:"rangeStart"`
	RangeEnd         string             `json:"rangeEnd"`
	TotalsByCategory map[string]float64 `json:"totalsByCategory"`
}

// ChatQueryRequest is the input for the chat query endpoint.
type ChatQueryRequest struct {
	Question     string        `json:"question"`
	LedgerSummary LedgerSummary `json:"ledgerSummary"`
}

// MetricCitation provides referenced metrics in chat responses.
type MetricCitation struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// ChatQueryResponse is the output from the chat query endpoint.
type ChatQueryResponse struct {
	Answer    string           `json:"answer"`
	Citations []MetricCitation `json:"citations"`
}

// ErrorResponse represents an API error response.
type ErrorResponse struct {
	Error string `json:"error"`
}
