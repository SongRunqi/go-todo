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

// OpenAIClient implements the Client interface for OpenAI API
type OpenAIClient struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(baseURL, apiKey, model string) *OpenAIClient {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/chat/completions"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}

	return &OpenAIClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		Client:  &http.Client{},
	}
}

// Chat sends a chat request to the OpenAI API
func (c *OpenAIClient) Chat(ctx context.Context, messages []Message) (string, error) {
	logger.Debug("Sending chat request to OpenAI API")

	// Build request
	req := Request{
		Model:    c.Model,
		Messages: messages,
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

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

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
		logger.Warnf("OpenAI API returned status: %d, body: %s", resp.StatusCode, string(respBody))
		return "", fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		logger.ErrorWithErr(err, "Failed to parse response")
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		logger.Error("No choices in API response")
		return "", fmt.Errorf("no response from API")
	}

	content := apiResp.Choices[0].Message.Content
	logger.Debug("Successfully received response from OpenAI API")

	// Remove markdown code fences if present
	content = removeJSONTag(content)

	return content, nil
}
