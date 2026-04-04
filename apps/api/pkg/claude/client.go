package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is a minimal Anthropic Claude API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

// New creates a new Claude client.
func New(apiKey, model string) *Client {
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		model:      model,
	}
}

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// completionRequest is the body sent to /v1/messages.
type completionRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system,omitempty"`
}

// completionResponse is the body received from /v1/messages.
type completionResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Complete sends a single prompt to Claude and returns the assistant reply.
func (c *Client) Complete(ctx context.Context, system, userMsg string) (string, error) {
	reqBody := completionRequest{
		Model:     c.model,
		MaxTokens: 2048,
		Messages:  []Message{{Role: "user", Content: userMsg}},
		System:    system,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("claude.Complete marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("claude.Complete new request: %w", err)
	}
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("claude.Complete http: %w", err)
	}
	defer resp.Body.Close()

	var result completionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("claude.Complete decode: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("claude API error: %s — %s", result.Error.Type, result.Error.Message)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("claude.Complete: empty response")
	}
	return result.Content[0].Text, nil
}
