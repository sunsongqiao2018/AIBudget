// Package ai provides an HTTP client adapter for communicating with the
// ai-orchestrator service. It implements the domain.AIOrchestratorClient interface.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sunsongqiao2018/AIBudget/services/edge-api/internal/domain"
)

// Client implements domain.AIOrchestratorClient using HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new AI orchestrator client.
func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Categorize forwards transaction candidates to the ai-orchestrator
// and returns categorized results.
func (c *Client) Categorize(ctx context.Context, req domain.CategorizeRequest) (domain.CategorizeResponse, error) {
	url := c.baseURL + "/internal/categorize"

	body, err := json.Marshal(req)
	if err != nil {
		return domain.CategorizeResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return domain.CategorizeResponse{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return domain.CategorizeResponse{}, domain.CategorizeError{
			Message: "failed to contact AI orchestrator",
			Cause:   err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.CategorizeResponse{}, domain.CategorizeError{
			Message: fmt.Sprintf("AI orchestrator returned status %d", resp.StatusCode),
		}
	}

	var result domain.CategorizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return domain.CategorizeResponse{}, domain.CategorizeError{
			Message: "failed to decode response",
			Cause:   err,
		}
	}

	return result, nil
}

// ChatQuery forwards a user question with ledger context to the ai-orchestrator
// and returns the chat response.
func (c *Client) ChatQuery(ctx context.Context, req domain.ChatQueryRequest) (domain.ChatQueryResponse, error) {
	url := c.baseURL + "/internal/chat/query"

	body, err := json.Marshal(req)
	if err != nil {
		return domain.ChatQueryResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return domain.ChatQueryResponse{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return domain.ChatQueryResponse{}, domain.ChatQueryError{
			Message: "failed to contact AI orchestrator",
			Cause:   err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.ChatQueryResponse{}, domain.ChatQueryError{
			Message: fmt.Sprintf("AI orchestrator returned status %d", resp.StatusCode),
		}
	}

	var result domain.ChatQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return domain.ChatQueryResponse{}, domain.ChatQueryError{
			Message: "failed to decode response",
			Cause:   err,
		}
	}

	return result, nil
}
