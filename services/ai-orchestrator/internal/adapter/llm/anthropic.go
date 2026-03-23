// Package llm provides LLM provider adapters for the ai-orchestrator service.
package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/sunsongqiao2018/AIBudget/services/ai-orchestrator/internal/domain"
)

const defaultModel = "claude-sonnet-4-6"
const defaultTimeoutSecs = 30

// AnthropicProvider implements domain.LLMProvider using the Anthropic Claude API.
type AnthropicProvider struct {
	client  anthropic.Client
	model   string
	timeout time.Duration
}

// NewAnthropicProvider creates a new AnthropicProvider from the given config.
// If cfg.Model is empty, it defaults to claude-sonnet-4-6.
// If cfg.TimeoutSecs is zero, it defaults to 30 seconds.
func NewAnthropicProvider(cfg domain.ProviderConfig) *AnthropicProvider {
	model := cfg.Model
	if model == "" {
		model = defaultModel
	}

	timeoutSecs := cfg.TimeoutSecs
	if timeoutSecs == 0 {
		timeoutSecs = defaultTimeoutSecs
	}

	return &AnthropicProvider{
		client:  anthropic.NewClient(option.WithAPIKey(cfg.APIKey)),
		model:   model,
		timeout: time.Duration(timeoutSecs) * time.Second,
	}
}

// categorizationResult is the JSON structure Claude is asked to return per transaction.
type categorizationResult struct {
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	Rationale  string  `json:"rationale"`
}

// Categorize sends transaction candidates to Claude and parses the categorization response.
func (p *AnthropicProvider) Categorize(ctx context.Context, candidates []domain.TransactionCandidate) ([]domain.CategorizedTransaction, error) {
	if len(candidates) == 0 {
		return []domain.CategorizedTransaction{}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	categoriesJSON, err := json.Marshal(domain.DefaultCategories())
	if err != nil {
		return nil, &domain.LLMError{Op: "categorize", Message: "failed to marshal categories", Err: domain.ErrCategorizationFailed}
	}

	candidatesJSON, err := json.Marshal(candidates)
	if err != nil {
		return nil, &domain.LLMError{Op: "categorize", Message: "failed to marshal candidates", Err: domain.ErrCategorizationFailed}
	}

	prompt := fmt.Sprintf(`You are a financial transaction categorizer.

Categorize each transaction in the provided list. Use ONLY the categories listed below.

Categories: %s

Transactions: %s

Respond with a valid JSON array containing exactly one object per transaction, in the same order.
Each object must have these fields:
- "category": string (must be one of the categories above)
- "confidence": number between 0.0 and 1.0
- "rationale": string (brief explanation)

Return ONLY the JSON array with no additional text, markdown, or code fences.`,
		string(categoriesJSON),
		string(candidatesJSON),
	)

	msg, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, p.mapError("categorize", err)
	}

	rawText := extractText(msg)
	rawText = strings.TrimSpace(rawText)

	var results []categorizationResult
	if err := json.Unmarshal([]byte(rawText), &results); err != nil {
		return nil, &domain.LLMError{
			Op:      "categorize",
			Message: fmt.Sprintf("failed to parse response JSON: %v", err),
			Err:     domain.ErrCategorizationFailed,
		}
	}

	if len(results) != len(candidates) {
		return nil, &domain.LLMError{
			Op:      "categorize",
			Message: fmt.Sprintf("expected %d results, got %d", len(candidates), len(results)),
			Err:     domain.ErrCategorizationFailed,
		}
	}

	categorized := make([]domain.CategorizedTransaction, len(candidates))
	for i, candidate := range candidates {
		categorized[i] = domain.CategorizedTransaction{
			TransactionCandidate: candidate,
			Category:             results[i].Category,
			Confidence:           results[i].Confidence,
			Rationale:            results[i].Rationale,
		}
	}

	return categorized, nil
}

// chatLLMResponse is the JSON structure Claude is asked to return for chat queries.
type chatLLMResponse struct {
	Answer    string `json:"answer"`
	Citations []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"citations"`
}

// ChatQuery sends a user question with the ledger context to Claude and returns the answer.
func (p *AnthropicProvider) ChatQuery(ctx context.Context, question string, ledger domain.LedgerSummary) (domain.ChatResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	ledgerJSON, err := json.Marshal(ledger)
	if err != nil {
		return domain.ChatResponse{}, &domain.LLMError{Op: "chat", Message: "failed to marshal ledger", Err: domain.ErrChatQueryFailed}
	}

	prompt := fmt.Sprintf(`You are a helpful personal finance assistant. Answer the user's question using the spending data provided.

Ledger summary: %s

User question: %s

Respond with a valid JSON object with these fields:
- "answer": string (natural language answer to the question)
- "citations": array of objects with "label" and "value" fields, citing specific data points that support your answer

Return ONLY the JSON object with no additional text, markdown, or code fences.`,
		string(ledgerJSON),
		question,
	)

	msg, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return domain.ChatResponse{}, p.mapError("chat", err)
	}

	rawText := strings.TrimSpace(extractText(msg))

	var llmResp chatLLMResponse
	if err := json.Unmarshal([]byte(rawText), &llmResp); err != nil {
		return domain.ChatResponse{}, &domain.LLMError{
			Op:      "chat",
			Message: fmt.Sprintf("failed to parse response JSON: %v", err),
			Err:     domain.ErrChatQueryFailed,
		}
	}

	citations := make([]domain.MetricCitation, len(llmResp.Citations))
	for i, c := range llmResp.Citations {
		citations[i] = domain.MetricCitation{Label: c.Label, Value: c.Value}
	}

	return domain.ChatResponse{
		Answer:    llmResp.Answer,
		Citations: citations,
	}, nil
}

// extractText pulls the plain text content from a Claude message response.
func extractText(msg *anthropic.Message) string {
	var sb strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return sb.String()
}

// mapError converts Anthropic SDK errors to domain sentinel errors.
func (p *AnthropicProvider) mapError(op string, err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return &domain.LLMError{Op: op, Message: "request timed out", Err: domain.ErrProviderTimeout}
	}

	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline"):
		return &domain.LLMError{Op: op, Message: errMsg, Err: domain.ErrProviderTimeout}
	case strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "unavailable") ||
		strings.Contains(errMsg, "overloaded") || strings.Contains(errMsg, "529"):
		return &domain.LLMError{Op: op, Message: errMsg, Err: domain.ErrProviderUnavailable}
	case op == "categorize":
		return &domain.LLMError{Op: op, Message: errMsg, Err: domain.ErrCategorizationFailed}
	default:
		return &domain.LLMError{Op: op, Message: errMsg, Err: domain.ErrChatQueryFailed}
	}
}
