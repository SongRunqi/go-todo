package app

import (
	"context"
	"os"
	"strings"

	"github.com/SongRunqi/go-todo/internal/ai"
	"github.com/SongRunqi/go-todo/internal/logger"
)

var (
	// aiClient is a package-level variable that can be overridden for testing
	aiClient ai.Client
)

// GetAIClient returns the AI client, initializing it if necessary
func GetAIClient() ai.Client {
	if aiClient == nil {
		provider := os.Getenv("AI_PROVIDER")
		baseURL := os.Getenv("LLM_BASE_URL")
		apiKey := os.Getenv("API_KEY")
		model := os.Getenv("LLM_MODEL")

		var err error
		aiClient, err = ai.NewClient(provider, baseURL, apiKey, model)
		if err != nil {
			logger.Warnf("Failed to create AI client: %v, falling back to DeepSeek", err)
			aiClient = ai.NewDeepSeekClient(baseURL, apiKey, model)
		}
	}
	return aiClient
}

// SetAIClient allows setting a custom AI client (useful for testing)
func SetAIClient(client ai.Client) {
	aiClient = client
}

func Chat(req OpenAIRequest) (string, error) {
	// Convert OpenAIRequest to ai.Message format
	messages := make([]ai.Message, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = ai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Use the AI client
	client := GetAIClient()
	ctx := context.Background()

	response, err := client.Chat(ctx, messages)
	if err != nil {
		logger.ErrorWithErr(err, "AI chat request failed")
		return "", err
	}

	logger.Debug("Successfully received response from AI")
	return response, nil
}

func removeJsonTag(str string) string {
	s := strings.Replace(str, "```json", "", 1)
	s = strings.Replace(s, "```", "", 1)
	return strings.TrimSpace(s)

}
