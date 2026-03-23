// Package llm provides LLM provider adapters for the ai-orchestrator service.
package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/domain"
)

// MockProvider is a deterministic LLM provider for testing and development.
// It implements domain.LLMProvider with rule-based responses.
type MockProvider struct {
	categories []string
}

// NewMockProvider creates a new mock LLM provider with default categories.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		categories: domain.DefaultCategories(),
	}
}

// NewMockProviderWithCategories creates a mock with custom categories.
func NewMockProviderWithCategories(categories []string) *MockProvider {
	return &MockProvider{
		categories: categories,
	}
}

// Categorize returns deterministic categories based on merchant name keywords.
func (m *MockProvider) Categorize(ctx context.Context, candidates []domain.TransactionCandidate) ([]domain.CategorizedTransaction, error) {
	results := make([]domain.CategorizedTransaction, len(candidates))

	for i, candidate := range candidates {
		category, confidence, rationale := m.categorizeByMerchant(candidate.Merchant)

		results[i] = domain.CategorizedTransaction{
			TransactionCandidate: candidate,
			Category:             category,
			Confidence:           confidence,
			Rationale:            rationale,
		}
	}

	return results, nil
}

func (m *MockProvider) categorizeByMerchant(merchant string) (category string, confidence float64, rationale string) {
	lower := strings.ToLower(merchant)

	switch {
	case strings.Contains(lower, "starbucks"), strings.Contains(lower, "coffee"),
		strings.Contains(lower, "restaurant"), strings.Contains(lower, "cafe"),
		strings.Contains(lower, "mcdonald"), strings.Contains(lower, "burger"),
		strings.Contains(lower, "pizza"), strings.Contains(lower, "sushi"):
		return "Food & Dining", 0.92, "Merchant name indicates food/dining establishment"

	case strings.Contains(lower, "uber"), strings.Contains(lower, "lyft"),
		strings.Contains(lower, "taxi"), strings.Contains(lower, "transit"),
		strings.Contains(lower, "gas"), strings.Contains(lower, "shell"),
		strings.Contains(lower, "parking"):
		return "Transportation", 0.88, "Merchant indicates transportation service"

	case strings.Contains(lower, "amazon"), strings.Contains(lower, "walmart"),
		strings.Contains(lower, "target"), strings.Contains(lower, "costco"),
		strings.Contains(lower, "store"), strings.Contains(lower, "shop"):
		return "Shopping", 0.85, "Merchant appears to be retail/shopping"

	case strings.Contains(lower, "netflix"), strings.Contains(lower, "spotify"),
		strings.Contains(lower, "cinema"), strings.Contains(lower, "theater"),
		strings.Contains(lower, "game"), strings.Contains(lower, "entertainment"):
		return "Entertainment", 0.87, "Merchant indicates entertainment service"

	case strings.Contains(lower, "electric"), strings.Contains(lower, "water"),
		strings.Contains(lower, "gas"), strings.Contains(lower, "internet"),
		strings.Contains(lower, "phone"), strings.Contains(lower, "utility"):
		return "Bills & Utilities", 0.90, "Merchant appears to be utility provider"

	case strings.Contains(lower, "pharmacy"), strings.Contains(lower, "doctor"),
		strings.Contains(lower, "hospital"), strings.Contains(lower, "gym"),
		strings.Contains(lower, "fitness"), strings.Contains(lower, "health"):
		return "Health & Wellness", 0.86, "Merchant indicates health/wellness service"

	case strings.Contains(lower, "hotel"), strings.Contains(lower, "airline"),
		strings.Contains(lower, "booking"), strings.Contains(lower, "travel"),
		strings.Contains(lower, "airbnb"):
		return "Travel", 0.91, "Merchant indicates travel-related service"

	case strings.Contains(lower, "course"), strings.Contains(lower, "university"),
		strings.Contains(lower, "school"), strings.Contains(lower, "book"),
		strings.Contains(lower, "education"):
		return "Education", 0.84, "Merchant appears education-related"

	default:
		return "Other", 0.70, fmt.Sprintf("Unable to categorize '%s' from merchant name", merchant)
	}
}

// ChatQuery returns deterministic answers based on keywords in the question.
func (m *MockProvider) ChatQuery(ctx context.Context, question string, ledger domain.LedgerSummary) (domain.ChatResponse, error) {
	lower := strings.ToLower(question)

	// Calculate total spending
	var total float64
	for _, amount := range ledger.TotalsByCategory {
		total += amount
	}

	// Check for category-specific questions
	for category, amount := range ledger.TotalsByCategory {
		lowerCategory := strings.ToLower(category)
		if strings.Contains(lower, lowerCategory) || strings.Contains(lower, strings.Split(lowerCategory, " & ")[0]) {
			return domain.ChatResponse{
				Answer: fmt.Sprintf("You spent $%.2f on %s between %s and %s.",
					amount, category, ledger.RangeStart, ledger.RangeEnd),
				Citations: []domain.MetricCitation{
					{Label: "Category", Value: category},
					{Label: "Amount", Value: fmt.Sprintf("$%.2f", amount)},
					{Label: "Period", Value: fmt.Sprintf("%s to %s", ledger.RangeStart, ledger.RangeEnd)},
				},
			}, nil
		}
	}

	// Check for total spending questions
	if strings.Contains(lower, "total") || strings.Contains(lower, "overall") ||
		strings.Contains(lower, "all") || strings.Contains(lower, "everything") {
		return domain.ChatResponse{
			Answer: fmt.Sprintf("Your total spending was $%.2f between %s and %s.",
				total, ledger.RangeStart, ledger.RangeEnd),
			Citations: []domain.MetricCitation{
				{Label: "Total Spending", Value: fmt.Sprintf("$%.2f", total)},
				{Label: "Period", Value: fmt.Sprintf("%s to %s", ledger.RangeStart, ledger.RangeEnd)},
			},
		}, nil
	}

	// Check for highest spending category
	if strings.Contains(lower, "most") || strings.Contains(lower, "highest") ||
		strings.Contains(lower, "largest") || strings.Contains(lower, "top") {
		var maxCategory string
		var maxAmount float64
		for category, amount := range ledger.TotalsByCategory {
			if amount > maxAmount {
				maxAmount = amount
				maxCategory = category
			}
		}

		return domain.ChatResponse{
			Answer: fmt.Sprintf("Your highest spending category was %s with $%.2f.",
				maxCategory, maxAmount),
			Citations: []domain.MetricCitation{
				{Label: "Top Category", Value: maxCategory},
				{Label: "Amount", Value: fmt.Sprintf("$%.2f", maxAmount)},
				{Label: "Period", Value: fmt.Sprintf("%s to %s", ledger.RangeStart, ledger.RangeEnd)},
			},
		}, nil
	}

	// Default response
	return domain.ChatResponse{
		Answer: fmt.Sprintf("I analyzed your spending of $%.2f across %d categories between %s and %s.",
			total, len(ledger.TotalsByCategory), ledger.RangeStart, ledger.RangeEnd),
		Citations: []domain.MetricCitation{
			{Label: "Total Spending", Value: fmt.Sprintf("$%.2f", total)},
			{Label: "Categories", Value: fmt.Sprintf("%d", len(ledger.TotalsByCategory))},
			{Label: "Period", Value: fmt.Sprintf("%s to %s", ledger.RangeStart, ledger.RangeEnd)},
		},
	}, nil
}
