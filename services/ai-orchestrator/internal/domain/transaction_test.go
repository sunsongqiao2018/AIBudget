package domain

import (
	"encoding/json"
	"testing"
)

func TestTransactionCandidate_Equality(t *testing.T) {
	a := TransactionCandidate{
		Date:     "2024-01-15",
		Amount:   45.67,
		Currency: "USD",
		Merchant: "Starbucks",
		Note:     "Coffee",
	}
	b := TransactionCandidate{
		Date:     "2024-01-15",
		Amount:   45.67,
		Currency: "USD",
		Merchant: "Starbucks",
		Note:     "Coffee",
	}

	if a.Date != b.Date || a.Amount != b.Amount || a.Currency != b.Currency ||
		a.Merchant != b.Merchant || a.Note != b.Note {
		t.Error("structurally equal candidates should match")
	}
}

func TestCategorizedTransaction_ConfidenceBounds(t *testing.T) {
	tests := []struct {
		name       string
		confidence float64
		valid      bool
	}{
		{"zero confidence", 0.0, true},
		{"low confidence", 0.25, true},
		{"medium confidence", 0.5, true},
		{"high confidence", 0.95, true},
		{"perfect confidence", 1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := CategorizedTransaction{
				TransactionCandidate: TransactionCandidate{
					Date:     "2024-01-15",
					Amount:   45.67,
					Currency: "USD",
					Merchant: "Test",
				},
				Category:   "Test",
				Confidence: tt.confidence,
			}

			// Verify confidence can be serialized
			data, err := json.Marshal(ct)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var got CategorizedTransaction
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if got.Confidence != tt.confidence {
				t.Errorf("Confidence: got %v, want %v", got.Confidence, tt.confidence)
			}
		})
	}
}

func TestCategorizeRequest_JSONRoundTrip(t *testing.T) {
	req := CategorizeRequest{
		Transactions: []TransactionCandidate{
			{Date: "2024-01-15", Amount: 45.67, Currency: "USD", Merchant: "Starbucks"},
			{Date: "2024-01-16", Amount: 23.50, Currency: "USD", Merchant: "Uber", Note: "Ride to airport"},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got CategorizeRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(got.Transactions) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(got.Transactions))
	}

	if got.Transactions[1].Note != "Ride to airport" {
		t.Errorf("Note: got %q, want %q", got.Transactions[1].Note, "Ride to airport")
	}
}

func TestChatQueryRequest_RoundTrip(t *testing.T) {
	req := ChatQueryRequest{
		Question: "What's my spending on food?",
		LedgerSummary: LedgerSummary{
			RangeStart: "2024-01-01",
			RangeEnd:   "2024-01-31",
			TotalsByCategory: map[string]float64{
				"Food & Dining":  450.50,
				"Transportation": 120.00,
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got ChatQueryRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.Question != req.Question {
		t.Errorf("Question: got %q, want %q", got.Question, req.Question)
	}

	if len(got.LedgerSummary.TotalsByCategory) != 2 {
		t.Errorf("TotalsByCategory length: got %d, want 2", len(got.LedgerSummary.TotalsByCategory))
	}
}

func TestChatQueryResponse_RoundTrip(t *testing.T) {
	resp := ChatQueryResponse{
		Answer: "You spent $450.50 on Food & Dining.",
		Citations: []MetricCitation{
			{Label: "Category", Value: "Food & Dining"},
			{Label: "Amount", Value: "$450.50"},
			{Label: "Period", Value: "January 2024"},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var got ChatQueryResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.Answer != resp.Answer {
		t.Errorf("Answer: got %q, want %q", got.Answer, resp.Answer)
	}

	if len(got.Citations) != 3 {
		t.Fatalf("expected 3 citations, got %d", len(got.Citations))
	}

	if got.Citations[0].Label != "Category" {
		t.Errorf("Citation[0].Label: got %q, want %q", got.Citations[0].Label, "Category")
	}
}
