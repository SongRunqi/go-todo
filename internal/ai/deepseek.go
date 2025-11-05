package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/SongRunqi/go-todo/internal/logger"
)

// DeepSeekClient implements the Client interface for DeepSeek API
type DeepSeekClient struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

// NewDeepSeekClient creates a new DeepSeek client
func NewDeepSeekClient(baseURL, apiKey, model string) *DeepSeekClient {
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/chat/completions"
	}
	if model == "" {
		model = "deepseek-chat"
	}

	return &DeepSeekClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
		Client:  &http.Client{},
	}
}

// Request represents the API request structure
type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Chat sends a chat request to the DeepSeek API
func (c *DeepSeekClient) Chat(ctx context.Context, messages []Message) (string, error) {
	logger.Debug("Sending chat request to DeepSeek API")

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
	httpReq.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.Client.Do(httpReq)
	if err != nil {
		logger.ErrorWithErr(err, "HTTP request failed")
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		logger.Warnf("API returned status: %d", resp.StatusCode)
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorWithErr(err, "Failed to read response body")
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	logger.Debugf("Raw API response: %s", string(respBody))

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
	logger.Debug("Successfully received response from API")

	// Remove markdown code fences if present
	content = removeJSONTag(content)

	return content, nil
}

// removeJSONTag removes markdown code fence formatting from responses
func removeJSONTag(str string) string {
	s := strings.Replace(str, "```json", "", 1)
	s = strings.Replace(s, "```", "", 1)
	return strings.TrimSpace(s)
}
