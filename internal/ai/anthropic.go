package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/SongRunqi/go-todo/internal/logger"
)

// AnthropicClient implements the Client interface for Anthropic Claude API
type AnthropicClient struct {
	BaseURL    string
	APIKey     string
	Model      string
	MaxTokens  int
	APIVersion string
	Client     *http.Client
}

// AnthropicRequest represents the request structure for Anthropic API
type AnthropicRequest struct {
	Model     string            `json:"model"`
	MaxTokens int               `json:"max_tokens"`
	System    string            `json:"system,omitempty"`
	Messages  []AnthropicMessage `json:"messages"`
}

// AnthropicMessage represents a message in Anthropic format
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents the response structure from Anthropic API
type AnthropicResponse struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Role         string                 `json:"role"`
	Content      []AnthropicContent     `json:"content"`
	Model        string                 `json:"model"`
	StopReason   string                 `json:"stop_reason"`
	StopSequence *string                `json:"stop_sequence"`
	Usage        AnthropicUsage         `json:"usage"`
	Error        *AnthropicError        `json:"error,omitempty"`
}

// AnthropicContent represents content blocks in Anthropic response
type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// AnthropicUsage represents token usage in Anthropic response
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicError represents an error from the Anthropic API
type AnthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// NewAnthropicClient creates a new Anthropic Claude client
func NewAnthropicClient(baseURL, apiKey, model string) *AnthropicClient {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1/messages"
	}
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	return &AnthropicClient{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		Model:      model,
		MaxTokens:  4096,
		APIVersion: "2023-06-01",
		Client:     &http.Client{},
	}
}

// Chat sends a chat request to the Anthropic Claude API
func (c *AnthropicClient) Chat(ctx context.Context, messages []Message) (string, error) {
	logger.Debug("Sending chat request to Anthropic Claude API")

	// Convert messages to Anthropic format
	// Anthropic requires system messages to be separate
	var systemPrompt string
	anthropicMessages := make([]AnthropicMessage, 0, len(messages))

	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			continue
		}
		anthropicMessages = append(anthropicMessages, AnthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Build request
	req := AnthropicRequest{
		Model:     c.Model,
		MaxTokens: c.MaxTokens,
		System:    systemPrompt,
		Messages:  anthropicMessages,
	}

	// Marshal to JSON
	body, err := json.Marshal(req)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to marshal request")
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(body))
	if err != nil {
		logger.ErrorWithErr(err, "Failed to create HTTP request")
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers - Anthropic uses different auth header
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)
	httpReq.Header.Set("anthropic-version", c.APIVersion)

	// Send request
	resp, err := c.Client.Do(httpReq)
	if err != nil {
		logger.ErrorWithErr(err, "HTTP request failed")
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to read response body")
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	logger.Debugf("Raw API response: %s", string(respBody))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		logger.Warnf("Anthropic API returned status: %d, body: %s", resp.StatusCode, string(respBody))
		return "", fmt.Errorf("Anthropic API error: status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp AnthropicResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		logger.ErrorWithErr(err, "Failed to parse response")
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error in response
	if apiResp.Error != nil {
		logger.Errorf("Anthropic API error: %s - %s", apiResp.Error.Type, apiResp.Error.Message)
		return "", fmt.Errorf("Anthropic API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Content) == 0 {
		logger.Error("No content in API response")
		return "", fmt.Errorf("no response from API")
	}

	// Extract text from content blocks
	var content string
	for _, block := range apiResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	logger.Debug("Successfully received response from Anthropic Claude API")

	// Remove markdown code fences if present
	content = removeJSONTag(content)

	return content, nil
}
