package domain

import (
	"encoding/json"
	"testing"
)

func TestTransactionCandidate_JSONSerialization(t *testing.T) {
	candidate := TransactionCandidate{
		Date:     "2024-01-15",
		Amount:   45.67,
		Currency: "USD",
		Merchant: "Starbucks",
		Note:     "Morning coffee",
	}

	data, err := json.Marshal(candidate)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var got TransactionCandidate
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if got.Date != candidate.Date {
		t.Errorf("Date: got %q, want %q", got.Date, candidate.Date)
	}
	if got.Amount != candidate.Amount {
		t.Errorf("Amount: got %v, want %v", got.Amount, candidate.Amount)
	}
	if got.Currency != candidate.Currency {
		t.Errorf("Currency: got %q, want %q", got.Currency, candidate.Currency)
	}
	if got.Merchant != candidate.Merchant {
		t.Errorf("Merchant: got %q, want %q", got.Merchant, candidate.Merchant)
	}
	if got.Note != candidate.Note {
		t.Errorf("Note: got %q, want %q", got.Note, candidate.Note)
	}
}

func TestCategorizedTransaction_JSONSerialization(t *testing.T) {
	ct := CategorizedTransaction{
		TransactionCandidate: TransactionCandidate{
			Date:     "2024-01-15",
			Amount:   45.67,
			Currency: "USD",
			Merchant: "Starbucks",
		},
		Category:   "Food & Dining",
		Confidence: 0.95,
		Rationale:  "Coffee shop purchase",
	}

	data, err := json.Marshal(ct)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var got CategorizedTransaction
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if got.Category != ct.Category {
		t.Errorf("Category: got %q, want %q", got.Category, ct.Category)
	}
	if got.Confidence != ct.Confidence {
		t.Errorf("Confidence: got %v, want %v", got.Confidence, ct.Confidence)
	}
	if got.Rationale != ct.Rationale {
		t.Errorf("Rationale: got %q, want %q", got.Rationale, ct.Rationale)
	}
}

func TestCategorizeRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CategorizeRequest
		wantErr bool
	}{
		{
			name: "valid with one transaction",
			req: CategorizeRequest{
				Transactions: []TransactionCandidate{
					{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid with multiple transactions",
			req: CategorizeRequest{
				Transactions: []TransactionCandidate{
					{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
					{Date: "2024-01-16", Amount: 23.50, Currency: "USD", Merchant: "Uber"},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty transactions is invalid for business logic",
			req:     CategorizeRequest{Transactions: []TransactionCandidate{}},
			wantErr: false, // structurally valid, but business layer should reject
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var got CategorizeRequest
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if len(got.Transactions) != len(tt.req.Transactions) {
				t.Errorf("transaction count: got %d, want %d", len(got.Transactions), len(tt.req.Transactions))
			}
		})
	}
}

func TestChatQueryRequest_Validation(t *testing.T) {
	tests := []struct {
		name string
		req  ChatQueryRequest
	}{
		{
			name: "valid request",
			req: ChatQueryRequest{
				Question: "What's my spending last month?",
				LedgerSummary: LedgerSummary{
					RangeStart: "2024-01-01",
					RangeEnd:   "2024-01-31",
					TotalsByCategory: map[string]float64{
						"Food & Dining": 450.50,
						"Transportation": 120.00,
					},
				},
			},
		},
		{
			name: "empty categories allowed",
			req: ChatQueryRequest{
				Question:      "Any spending?",
				LedgerSummary: LedgerSummary{
					RangeStart:       "2024-01-01",
					RangeEnd:         "2024-01-31",
					TotalsByCategory: map[string]float64{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.req)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var got ChatQueryRequest
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if got.Question != tt.req.Question {
				t.Errorf("Question: got %q, want %q", got.Question, tt.req.Question)
			}
			if got.LedgerSummary.RangeStart != tt.req.LedgerSummary.RangeStart {
				t.Errorf("RangeStart: got %q, want %q", got.LedgerSummary.RangeStart, tt.req.LedgerSummary.RangeStart)
			}
		})
	}
}

func TestLedgerSummary_JSONSerialization(t *testing.T) {
	summary := LedgerSummary{
		RangeStart: "2024-01-01",
		RangeEnd:   "2024-01-31",
		TotalsByCategory: map[string]float64{
			"Food & Dining":  450.50,
			"Transportation": 120.00,
			"Shopping":       89.99,
		},
	}

	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var got LedgerSummary
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if got.RangeStart != summary.RangeStart {
		t.Errorf("RangeStart: got %q, want %q", got.RangeStart, summary.RangeStart)
	}
	if got.RangeEnd != summary.RangeEnd {
		t.Errorf("RangeEnd: got %q, want %q", got.RangeEnd, summary.RangeEnd)
	}
	if len(got.TotalsByCategory) != len(summary.TotalsByCategory) {
		t.Errorf("TotalsByCategory length: got %d, want %d", len(got.TotalsByCategory), len(summary.TotalsByCategory))
	}
}

func TestErrorResponse_JSONSerialization(t *testing.T) {
	errResp := ErrorResponse{Error: "validation failed: missing merchant"}

	data, err := json.Marshal(errResp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var got ErrorResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if got.Error != errResp.Error {
		t.Errorf("Error: got %q, want %q", got.Error, errResp.Error)
	}
}
