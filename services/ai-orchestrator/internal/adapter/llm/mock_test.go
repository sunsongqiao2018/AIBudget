package llm

import (
	"context"
	"strings"
	"testing"

	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/domain"
)

func TestMockProvider_Categorize_Food(t *testing.T) {
	provider := NewMockProvider()

	candidates := []domain.TransactionCandidate{
		{Date: "2024-01-15", Amount: 5.67, Currency: "USD", Merchant: "Starbucks"},
		{Date: "2024-01-16", Amount: 12.50, Currency: "USD", Merchant: "Joe's Cafe"},
		{Date: "2024-01-17", Amount: 8.99, Currency: "USD", Merchant: "Pizza Hut"},
	}

	results, err := provider.Categorize(context.Background(), candidates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for i, result := range results {
		if result.Category != "Food & Dining" {
			t.Errorf("result[%d].Category: got %q, want %q", i, result.Category, "Food & Dining")
		}
		if result.Confidence < 0.9 {
			t.Errorf("result[%d].Confidence: got %v, want >= 0.9", i, result.Confidence)
		}
		if result.Rationale == "" {
			t.Errorf("result[%d].Rationale: should not be empty", i)
		}
	}
}

func TestMockProvider_Categorize_Transportation(t *testing.T) {
	provider := NewMockProvider()

	candidates := []domain.TransactionCandidate{
		{Date: "2024-01-15", Amount: 23.50, Currency: "USD", Merchant: "Uber"},
		{Date: "2024-01-16", Amount: 45.00, Currency: "USD", Merchant: "Shell Gas"},
	}

	results, err := provider.Categorize(context.Background(), candidates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, result := range results {
		if result.Category != "Transportation" {
			t.Errorf("result[%d].Category: got %q, want %q", i, result.Category, "Transportation")
		}
	}
}

func TestMockProvider_Categorize_Shopping(t *testing.T) {
	provider := NewMockProvider()

	candidates := []domain.TransactionCandidate{
		{Date: "2024-01-15", Amount: 89.99, Currency: "USD", Merchant: "Amazon"},
		{Date: "2024-01-16", Amount: 45.67, Currency: "USD", Merchant: "Target"},
	}

	results, err := provider.Categorize(context.Background(), candidates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, result := range results {
		if result.Category != "Shopping" {
			t.Errorf("result[%d].Category: got %q, want %q", i, result.Category, "Shopping")
		}
	}
}

func TestMockProvider_Categorize_Other(t *testing.T) {
	provider := NewMockProvider()

	candidates := []domain.TransactionCandidate{
		{Date: "2024-01-15", Amount: 100.00, Currency: "USD", Merchant: "Unknown Vendor XYZ"},
	}

	results, err := provider.Categorize(context.Background(), candidates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Category != "Other" {
		t.Errorf("Category: got %q, want %q", results[0].Category, "Other")
	}

	if results[0].Confidence > 0.75 {
		t.Errorf("Confidence for unknown should be low, got %v", results[0].Confidence)
	}
}

func TestMockProvider_Categorize_Empty(t *testing.T) {
	provider := NewMockProvider()

	results, err := provider.Categorize(context.Background(), []domain.TransactionCandidate{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestMockProvider_ChatQuery_TotalSpending(t *testing.T) {
	provider := NewMockProvider()

	ledger := domain.LedgerSummary{
		RangeStart: "2024-01-01",
		RangeEnd:   "2024-01-31",
		TotalsByCategory: map[string]float64{
			"Food & Dining":  450.50,
			"Transportation": 120.00,
			"Shopping":       89.99,
		},
	}

	questions := []string{
		"What's my total spending?",
		"How much did I spend overall?",
		"What was everything I spent?",
	}

	for _, question := range questions {
		t.Run(question, func(t *testing.T) {
			resp, err := provider.ChatQuery(context.Background(), question, ledger)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(resp.Answer, "660.49") {
				t.Errorf("Answer should contain total $660.49, got: %s", resp.Answer)
			}

			if len(resp.Citations) < 2 {
				t.Errorf("Expected at least 2 citations, got %d", len(resp.Citations))
			}
		})
	}
}

func TestMockProvider_ChatQuery_CategorySpecific(t *testing.T) {
	provider := NewMockProvider()

	ledger := domain.LedgerSummary{
		RangeStart: "2024-01-01",
		RangeEnd:   "2024-01-31",
		TotalsByCategory: map[string]float64{
			"Food & Dining":  450.50,
			"Transportation": 120.00,
		},
	}

	resp, err := provider.ChatQuery(context.Background(), "What did I spend on food?", ledger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(resp.Answer, "Food & Dining") {
		t.Errorf("Answer should mention 'Food & Dining', got: %s", resp.Answer)
	}

	if !strings.Contains(resp.Answer, "450.50") {
		t.Errorf("Answer should contain $450.50, got: %s", resp.Answer)
	}

	found := false
	for _, citation := range resp.Citations {
		if citation.Label == "Category" && citation.Value == "Food & Dining" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected citation for Food & Dining category")
	}
}

func TestMockProvider_ChatQuery_HighestSpending(t *testing.T) {
	provider := NewMockProvider()

	ledger := domain.LedgerSummary{
		RangeStart: "2024-01-01",
		RangeEnd:   "2024-01-31",
		TotalsByCategory: map[string]float64{
			"Food & Dining":  450.50,
			"Transportation": 120.00,
			"Shopping":       89.99,
		},
	}

	questions := []string{
		"What was my highest spending category?",
		"Where did I spend the most?",
		"What's my top category?",
	}

	for _, question := range questions {
		t.Run(question, func(t *testing.T) {
			resp, err := provider.ChatQuery(context.Background(), question, ledger)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(resp.Answer, "Food & Dining") {
				t.Errorf("Answer should mention 'Food & Dining' as highest, got: %s", resp.Answer)
			}

			if !strings.Contains(resp.Answer, "450.50") {
				t.Errorf("Answer should contain $450.50, got: %s", resp.Answer)
			}
		})
	}
}

func TestMockProvider_ChatQuery_Default(t *testing.T) {
	provider := NewMockProvider()

	ledger := domain.LedgerSummary{
		RangeStart: "2024-01-01",
		RangeEnd:   "2024-01-31",
		TotalsByCategory: map[string]float64{
			"Food & Dining": 100.00,
		},
	}

	resp, err := provider.ChatQuery(context.Background(), "Tell me something interesting", ledger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Answer == "" {
		t.Error("Answer should not be empty")
	}

	if len(resp.Citations) == 0 {
		t.Error("Expected at least one citation")
	}
}

func TestMockProviderWithCategories(t *testing.T) {
	customCategories := []string{"Business", "Personal", "Investment"}
	provider := NewMockProviderWithCategories(customCategories)

	if len(provider.categories) != 3 {
		t.Errorf("expected 3 custom categories, got %d", len(provider.categories))
	}

	if provider.categories[0] != "Business" {
		t.Errorf("expected first category to be Business, got %q", provider.categories[0])
	}
}
